package weather

import (
	"context"
	"weather-data-aggregator-service/src/domain/model"
)

// PostgresRepository represent repository contract
type PostgresRepository interface {
	GetCurrent(ctx context.Context, q model.CurrentQuery) (*model.AggregatedWeatherDataResp, error)
	GetForecast(ctx context.Context, q model.ForecastQuery) (*model.AggregatedForecast, error)
}
