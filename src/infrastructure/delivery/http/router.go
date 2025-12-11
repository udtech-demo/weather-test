package http

import (
	"fmt"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/viper"
	"os"
	"time"
	"weather-data-aggregator-service/src/registry"
)

func newLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "${time} ${status} ${method} ${path} (${remote_ip}) ${latency_human} ${req_header:Request-ID}\n",
		TimeFormat: "2006/01/02 15:04:05.000", // точность до миллисекунд
		TimeZone:   "Local",

		DisableColors: false,
		Output:        os.Stdout,
	})
}

func NewBase(f *fiber.App, c registry.APIController) {
	env := viper.GetString("env")

	f.Use(cors.New())

	f.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))

	if env == "local" {
		f.Use(logger.New(logger.Config{
			Format:        "${time} ${status} ${method} ${path} (${remote_ip}) ${latency_human}\n",
			TimeFormat:    "2006/01/02 15:04:05.000",
			TimeZone:      "Local",
			DisableColors: false,
			Output:        os.Stdout,
		}))

		f.Use(func(c *fiber.Ctx) error {
			start := time.Now()
			err := c.Next()
			duration := time.Since(start)
			requestID := c.Get("Request-ID") // читаем заголовок напрямую

			fmt.Fprintf(os.Stdout, "%s %d %s %s (%s) %s Request-ID: %s\n",
				time.Now().Format("2006/01/02 15:04:05.000"),
				c.Response().StatusCode(),
				c.Method(),
				c.Path(),
				c.IP(),
				duration,
				requestID,
			)

			return err
		})
	}

	f.Use(recover.New())

	f.Use(sentryfiber.New(sentryfiber.Options{Repanic: true}))
}

func NewFiberRouter(f *fiber.App, c registry.APIController) {
	NewBase(f, c)

	// Base routs
	apiV1 := f.Group("/api/v1")
	{
		apiV1.Get("/health", c.Weather.HealthCheck)
	}

	apiV1Weather := apiV1.Group("/weather")
	{
		apiV1Weather.Get("/current", c.Weather.GetCurrent)
		apiV1Weather.Get("/forecast", c.Weather.GetForecast)

	}
}
