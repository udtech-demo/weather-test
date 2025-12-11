package weather

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
	"github.com/uptrace/bun"
	"weather-data-aggregator-service/src/domain/model"
)

type WeatherClient struct {
	cities            []model.City
	dbClient          *bun.DB
	openWeatherAPIKey string
	weatherAPIKey     string
}

func createWeatherClient(dbClient *bun.DB) *WeatherClient {

	owKey := viper.GetString("open_weather.key")
	waKey := viper.GetString("weather_api.key")

	if owKey == "" || waKey == "" {
		log.Errorf("OpenWeather and WeatherAPI keys are required")
	}

	wc := &WeatherClient{
		dbClient:          dbClient,
		openWeatherAPIKey: owKey,
		weatherAPIKey:     waKey,
	}

	wc.LoadCitiesFromDB()
	return wc
}

func (w *WeatherClient) LoadCitiesFromDB() {
	var cities []model.City

	err := w.dbClient.NewSelect().
		Model(&cities).
		Where("enabled = ?", true).
		Scan(context.Background())

	if err != nil {
		log.Fatalf("Failed to load cities: %v", err)
	}

	w.cities = cities
	log.Errorf("Loaded %d cities from DB", len(cities))
}

func (w *WeatherClient) GetOpenWeatherAPIKeyKey() string {
	return w.openWeatherAPIKey
}

func (w *WeatherClient) GetWeatherAPIKey() string {
	return w.weatherAPIKey
}
