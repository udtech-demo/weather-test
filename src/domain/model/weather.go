package model

import (
	"github.com/google/uuid"
	"time"

	"github.com/uptrace/bun"
)

// WeatherData struct for normalized data
type WeatherData struct {
	bun.BaseModel `bun:"table:weather_data"`

	ID          uuid.UUID `json:"id" bun:",pk,nullzero,type:uuid,default:uuid_generate_v4()"`
	CityID      uuid.UUID `json:"city_id" bun:"city_id,notnull"`
	City        *City     `json:"city,omitempty" bun:"rel:belongs-to,join:city_id=id"`
	Source      string    `bun:"source,notnull"`
	Temperature float64   `bun:"temperature"`
	Humidity    int       `bun:"humidity"`
	WindSpeed   float64   `bun:"wind_speed"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

// AggregatedWeatherData struct for aggregated data
type AggregatedWeatherData struct {
	bun.BaseModel `bun:"table:aggregated_weather_data"`

	ID          uuid.UUID `json:"id" bun:",pk,nullzero,type:uuid,default:uuid_generate_v4()"`
	CityID      uuid.UUID `json:"city_id" bun:"city_id,notnull"`
	City        *City     `json:"city,omitempty" bun:"rel:belongs-to,join:city_id=id"`
	Temperature float64   `bun:"temperature"`
	Humidity    int       `bun:"humidity"`
	WindSpeed   float64   `bun:"wind_speed"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

type AggregatedWeatherDataResp struct {
	City        *City   `json:"city,omitempty" bun:"rel:belongs-to,join:city_id=id"`
	Temperature float64 `bun:"temperature"`
	Humidity    int     `bun:"humidity"`
	WindSpeed   float64 `bun:"wind_speed"`
}

type CurrentQuery struct {
	City string `query:"city"`
}

type ForecastQuery struct {
	City string `query:"city"`
	Days int    `query:"days"`
}

type ForecastDay struct {
	Date        time.Time `json:"date"`
	Temperature float64   `json:"temperature"`
	Humidity    int       `json:"humidity"`
	WindSpeed   float64   `json:"wind_speed"`
	Description string    `json:"description"`
}

type ForecastData struct {
	City string        `json:"city"`
	Days []ForecastDay `json:"days"`
}

type OneCallRespOWM struct {
	Daily []Daily `json:"daily"`
}

type Daily struct {
	Dt        int64     `json:"dt"`
	Temp      Temp      `json:"temp"`
	Humidity  int       `json:"humidity"`
	WindSpeed float64   `json:"wind_speed"`
	Weather   []Weather `json:"weather"`
}

type Temp struct {
	Day float64 `json:"day"`
}

type Weather struct {
	Description string `json:"description"`
}

type AggregatedForecastDay struct {
	Date         time.Time `json:"date"`
	Temperature  float64   `json:"temperature_avg"`
	Humidity     int       `json:"humidity_avg"`
	WindSpeed    float64   `json:"wind_speed_avg"`
	Descriptions []string  `json:"descriptions"`
}

type AggregatedForecast struct {
	City string                  `json:"city"`
	Days []AggregatedForecastDay `json:"days"`
}

type OneCallRespWA struct {
	Location Location `json:"location"`
	Forecast Forecast `json:"forecast"`
}

type Location struct {
	Name string `json:"name"`
}

type Forecast struct {
	Forecastday []ForecastDayWA `json:"forecastday"`
}

type ForecastDayWA struct {
	Date string `json:"date"`
	Day  Day    `json:"day"`
}

type Day struct {
	AvgtempC    float64   `json:"avgtemp_c"`
	Avghumidity float64   `json:"avghumidity"`
	MaxwindKph  float64   `json:"maxwind_kph"`
	Condition   Condition `json:"condition"`
}

type Condition struct {
	Text string `json:"text"`
}
