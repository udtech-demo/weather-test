package usecase

import (
	"context"
	"weather-data-aggregator-service/src/domain/model"
	"weather-data-aggregator-service/src/parts/weather"
)

type weatherUseCase struct {
	pRepo weather.PostgresRepository
}

func NewWeatherUseCase(pRepo weather.PostgresRepository) weather.UseCase {
	return &weatherUseCase{pRepo}
}

func (w *weatherUseCase) GetCurrent(ctx context.Context, q model.CurrentQuery) (*model.AggregatedWeatherDataResp, error) {
	return w.pRepo.GetCurrent(ctx, q)
}

func (w *weatherUseCase) GetForecast(ctx context.Context, q model.ForecastQuery) (*model.AggregatedForecast, error) {
	return w.pRepo.GetForecast(ctx, q)
}
