package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/uptrace/bun"
	"net/http"
	"net/url"
	"time"
	"weather-data-aggregator-service/src/domain/model"
	weather_client "weather-data-aggregator-service/src/infrastructure/weather"
	"weather-data-aggregator-service/src/parts/weather"
)

type weatherPostgresRepository struct {
	db            *bun.DB
	weatherClient *weather_client.WeatherClient
}

func NewWeatherPostgresRepository(db *bun.DB, weatherClient *weather_client.WeatherClient) weather.PostgresRepository {
	return &weatherPostgresRepository{db, weatherClient}
}
func (w *weatherPostgresRepository) GetCurrent(ctx context.Context, q model.CurrentQuery) (*model.AggregatedWeatherDataResp, error) {

	var city model.City
	err := w.db.NewSelect().Model(&city).Where("name = ?", q.City).Scan(ctx)
	if err != nil {
		return nil, err
	}

	var awd model.AggregatedWeatherData
	err = w.db.NewSelect().Model(&awd).Relation("City").Where("city_id = ?", city.ID).
		Order("created_at DESC").Limit(1).Scan(ctx)
	if err != nil {
		return nil, err
	}

	res := model.AggregatedWeatherDataResp{
		City:        awd.City,
		Temperature: awd.Temperature,
		Humidity:    awd.Humidity,
		WindSpeed:   awd.WindSpeed,
	}
	return &res, nil
}
func (w *weatherPostgresRepository) GetForecast(ctx context.Context, q model.ForecastQuery) (*model.AggregatedForecast, error) {
	escapedCity := url.QueryEscape(q.City)

	//openweathermap - paid service
	//owmData, err := fetchOWMForecast(escapedCity, w.weatherClient.GetOpenWeatherAPIKeyKey(), q.Days)
	//if err != nil {
	//	return nil, err
	//}

	apiData, err := fetchWeatherAPIForecast(escapedCity, w.weatherClient.GetWeatherAPIKey(), q.Days)
	if err != nil {
		return nil, err
	}

	//openweathermap - paid service
	//if len(owmData) != len(apiData) {
	//	return nil, fmt.Errorf("mismatch in days length between APIs")
	//}

	//aggregated := make([]model.AggregatedForecastDay, len(owmData))
	//for i := 0; i < len(owmData); i++ {
	//	aggregated[i] = model.AggregatedForecastDay{
	//		Date:         owmData[i].Date,
	//		Temperature:  (owmData[i].Temperature + apiData[i].Temperature) / 2,
	//		Humidity:     (owmData[i].Humidity + apiData[i].Humidity) / 2,
	//		WindSpeed:    (owmData[i].WindSpeed + apiData[i].WindSpeed) / 2,
	//		Descriptions: []string{owmData[i].Description, apiData[i].Description},
	//	}
	//}

	aggregated := make([]model.AggregatedForecastDay, len(apiData))
	for i := 0; i < len(apiData); i++ {
		aggregated[i] = model.AggregatedForecastDay{
			Date:         apiData[i].Date,
			Temperature:  apiData[i].Temperature,
			Humidity:     apiData[i].Humidity,
			WindSpeed:    apiData[i].WindSpeed,
			Descriptions: []string{apiData[i].Description},
		}
	}

	return &model.AggregatedForecast{
		City: q.City,
		Days: aggregated,
	}, nil
}

func fetchOWMForecast(cityName, apiKey string, days int) ([]model.ForecastDay, error) {
	if days < 1 || days > 7 {
		return nil, fmt.Errorf("days must be between 1 and 7")
	}

	cityURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", cityName, apiKey)
	resp, err := http.Get(cityURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var cityResp model.СityResp
	if err := json.NewDecoder(resp.Body).Decode(&cityResp); err != nil {
		return nil, err
	}

	oneCallURL := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/onecall?lat=%f&lon=%f&exclude=minutely,hourly,alerts,current&units=metric&appid=%s",
		cityResp.Coord.Lat, cityResp.Coord.Lon, apiKey,
	)

	resp2, err := http.Get(oneCallURL)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var oneCallResp model.OneCallRespOWM
	if err := json.NewDecoder(resp2.Body).Decode(&oneCallResp); err != nil {
		return nil, err
	}

	if len(oneCallResp.Daily) < days {
		days = len(oneCallResp.Daily)
	}

	result := make([]model.ForecastDay, days)
	for i := 0; i < days; i++ {
		d := oneCallResp.Daily[i]
		desc := ""
		if len(d.Weather) > 0 {
			desc = d.Weather[0].Description
		}
		result[i] = model.ForecastDay{
			Date:        time.Unix(d.Dt, 0),
			Temperature: d.Temp.Day,
			Humidity:    d.Humidity,
			WindSpeed:   d.WindSpeed,
			Description: desc,
		}
	}

	return result, nil
}

func fetchWeatherAPIForecast(cityName, apiKey string, days int) ([]model.ForecastDay, error) {
	if days < 1 || days > 7 {
		return nil, fmt.Errorf("days must be between 1 and 7")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&days=%d", apiKey, cityName, days)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp model.OneCallRespWA

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	result := make([]model.ForecastDay, len(apiResp.Forecast.Forecastday))
	for i, d := range apiResp.Forecast.Forecastday {
		date, _ := time.Parse("2006-01-02", d.Date)
		result[i] = model.ForecastDay{
			Date:        date,
			Temperature: d.Day.AvgtempC,
			Humidity:    int(d.Day.Avghumidity),
			WindSpeed:   d.Day.MaxwindKph / 3.6, // конвертируем km/h в m/s
			Description: d.Day.Condition.Text,
		}
	}

	return result, nil
}
