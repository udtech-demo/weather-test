package server

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"
	serverHttp "weather-data-aggregator-service/src/infrastructure/delivery/http"
	"weather-data-aggregator-service/src/infrastructure/storage/postgres"
	"weather-data-aggregator-service/src/infrastructure/storage/redis"
	"weather-data-aggregator-service/src/infrastructure/weather"
	"weather-data-aggregator-service/src/registry"
	scheduled_tasks "weather-data-aggregator-service/src/schedule"
)

type App struct {
	httpServer *http.Server
	f          *fiber.App
}

//func NewApp() *App {
//	// Set server timezone
//	location, err := time.LoadLocation(viper.GetString("server_timezone"))
//	if err != nil {
//		panic(err)
//	}
//	time.Local = location
//
//	file, _ := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
//	iw := io.MultiWriter(os.Stdout, file)
//	log.SetOutput(iw)
//	log.SetOutput(file)
//
//	// Init storages
//	db := postgres.InitPostgres()
//	rdb := redis.InitRedis()
//
//	weatherClient := weather.InitWeatherAPI(db)
//
//	// Register and create controller
//	apiController := registry.NewRegistry(db, rdb, weatherClient).NewAPIController()
//
//	// Initialize fiber instance
//	f := fiber.New(fiber.Config{
//		ErrorHandler: fiberErrorHandler,
//	})
//
//	// Register routs
//	serverHttp.NewFiberRouter(f, apiController)
//
//	scheduled_tasks.RunCronJobs(weatherClient)
//
//	// Start server
//	s := &http.Server{
//		Addr:           viper.GetString("http.port"),
//		ReadTimeout:    30 * time.Second,
//		WriteTimeout:   30 * time.Second,
//		MaxHeaderBytes: 1 << 20,
//	}
//
//	return &App{
//		httpServer: s,
//		f:          f,
//	}
//}

func NewApp() *App {
	// timezone
	location, err := time.LoadLocation(viper.GetString("server_timezone"))
	if err != nil {
		panic(err)
	}
	time.Local = location

	// logging
	file, _ := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	iw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(iw)

	db := postgres.InitPostgres()
	rdb := redis.InitRedis()
	weatherClient := weather.InitWeatherAPI(db)

	apiController := registry.NewRegistry(db, rdb, weatherClient).NewAPIController()

	f := fiber.New(fiber.Config{
		ErrorHandler: fiberErrorHandler,
	})

	serverHttp.NewFiberRouter(f, apiController)

	scheduled_tasks.RunCronJobs(weatherClient)

	return &App{
		f: f,
	}
}

func (a *App) Run() error {
	go func() {
		if err := a.f.Listen(viper.GetString("http.port")); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	return a.f.Shutdown()
}

//func (a *App) Run() error {
//	// Start server
//	go func() {
//		if err := a.f.Listen(a.httpServer.Addr); err != nil {
//			panic(err)
//		}
//	}()
//
//	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
//	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
//	quit := make(chan os.Signal, 1)
//	signal.Notify(quit, os.Interrupt)
//	<-quit
//
//	err := a.f.Shutdown()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return nil
//}

var fiberErrorHandler = func(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		err = errors.New(e.Message)
	}

	log.Errorf("[ERROR] %s %s - %v", c.Method(), c.OriginalURL(), err)

	return c.Status(code).JSON(fiber.Map{
		"error":  err.Error(),
		"status": code,
		"path":   c.OriginalURL(),
		"time":   time.Now().Format(time.RFC3339),
	})
}
