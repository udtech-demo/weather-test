package registry

import (
	"github.com/go-redis/redis/v8"
	"github.com/uptrace/bun"
	weather_client "weather-data-aggregator-service/src/infrastructure/weather"
	"weather-data-aggregator-service/src/parts/weather"
)

type APIController struct {
	Weather interface{ weather.Controller }
}

type register struct {
	db            *bun.DB
	rdb           *redis.Client
	weatherClient *weather_client.WeatherClient
}

type Registry interface {
	NewAPIController() APIController
}

func NewRegistry(db *bun.DB, rdb *redis.Client, weatherClient *weather_client.WeatherClient) Registry {
	return &register{db, rdb, weatherClient}
}

func (r *register) NewAPIController() APIController {
	return APIController{
		Weather: r.NewWeatherController(),
	}
}
