package registry

import (
	"weather-data-aggregator-service/src/parts/weather"
	"weather-data-aggregator-service/src/parts/weather/delivery/http"
	"weather-data-aggregator-service/src/parts/weather/repository/postgres"
	"weather-data-aggregator-service/src/parts/weather/usecase"
)

func (r *register) NewWeatherController() weather.Controller {
	return http.NewWeatherController(r.NewWeatherUseCase())
}

func (r *register) NewWeatherUseCase() weather.UseCase {
	return usecase.NewWeatherUseCase(r.NewWeatherPostgresRepository())
}

func (r *register) NewWeatherPostgresRepository() weather.PostgresRepository {
	return postgres.NewWeatherPostgresRepository(r.db, r.weatherClient)
}
