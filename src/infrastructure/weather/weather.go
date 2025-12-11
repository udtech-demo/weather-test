package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenk/backoff"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"github.com/uptrace/bun"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
	"weather-data-aggregator-service/src/domain/model"
)

// API Clients with circuit breakers
var (
	openWeatherCB = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "openweather",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < 10 {
				return false
			}
			failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRate > 0.5
		},
	})

	weatherAPICB = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "weatherapi",
		MaxRequests: 1,
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < 10 {
				return false
			}
			failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRate > 0.5
		},
	})
)

// Generic retry wrapper
func retry(attempts int, base time.Duration, fn func() error) error {
	var err error
	delay := base

	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(delay)
		delay *= 2
	}
	return err
}

func (w *WeatherClient) fetchWeatherData() {
	for _, city := range w.cities {
		w.fetchCityWeather(&city)
	}
}

func (w *WeatherClient) fetchCityWeather(city *model.City) {
	log.Info("Fetching weather for city: %s", city.Name)

	var wg sync.WaitGroup
	wg.Add(2)

	timeNow := time.Now()

	var (
		openWeatherData model.WeatherData
		weatherAPIData  model.WeatherData
	)

	// OpenWeather Fetch
	go func() {
		defer wg.Done()

		_, err := openWeatherCB.Execute(func() (interface{}, error) {
			return nil, retry(3, 300*time.Millisecond, func() error {
				return fetchOpenWeather(&openWeatherData, city, w.openWeatherAPIKey, timeNow)
			})
		})

		if err != nil {
			log.Errorf("[ERROR] OpenWeather fetch failed for %s (CB=%v): %v",
				city, openWeatherCB.State(), err)
			return
		}
	}()

	// WeatherAPI Fetch
	go func() {
		defer wg.Done()

		_, err := weatherAPICB.Execute(func() (interface{}, error) {
			return nil, retry(3, 300*time.Millisecond, func() error {
				return fetchWeatherAPI(&weatherAPIData, city, w.weatherAPIKey, timeNow)
			})
		})

		if err != nil {
			log.Errorf("[ERROR] WeatherAPI fetch failed for %s (CB=%v): %v",
				city.Name, weatherAPICB.State(), err)
			return
		}
	}()

	wg.Wait()

	if openWeatherData.CityID != uuid.Nil && weatherAPIData.CityID != uuid.Nil {
		ctx := context.Background()

		tx, err := w.dbClient.BeginTx(ctx, nil)
		if err != nil {
			log.Errorf("Failed to begin transaction: %v", err)
		}
		w.saveWeatherData(&tx, &openWeatherData)
		w.saveWeatherData(&tx, &weatherAPIData)

		w.saveAggregatedWeatherData(&tx, &model.AggregatedWeatherData{
			CityID:      city.ID,
			Temperature: (openWeatherData.Temperature + weatherAPIData.Temperature) / 2,
			Humidity:    (openWeatherData.Humidity + weatherAPIData.Humidity) / 2,
			WindSpeed:   math.Round(((openWeatherData.WindSpeed+weatherAPIData.WindSpeed)/2)*100) / 100,
		})

		if err := tx.Commit(); err != nil {
			log.Errorf("Failed to commit transaction: %v", err)
		}
	}
}

func fetchOpenWeather(data *model.WeatherData, city *model.City, openWeatherAPIKey string, timeNow time.Time) error {
	escapedCity := url.QueryEscape(city.Name)
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		escapedCity, openWeatherAPIKey,
	)

	var resp *http.Response

	operation := func() error {
		r, err := http.Get(url)
		if err != nil {
			return err
		}

		if r.StatusCode == http.StatusTooManyRequests {
			r.Body.Close()
			return backoff.Permanent(fmt.Errorf("rate limit exceeded"))
		}

		if r.StatusCode >= 400 {
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body.Close()
			return fmt.Errorf("HTTP %d: %s", r.StatusCode, string(bodyBytes))
		}

		resp = r
		return nil
	}

	if err := backoff.Retry(operation, backoff.NewExponentialBackOff()); err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	data.CityID = city.ID
	data.Source = "OpenWeatherMap"
	data.Temperature = result.Main.Temp
	data.Humidity = result.Main.Humidity
	data.WindSpeed = math.Round(result.Wind.Speed*100) / 100
	data.CreatedAt = timeNow

	return nil
}

func fetchWeatherAPI(data *model.WeatherData, city *model.City, weatherAPIKey string, timeNow time.Time) error {
	escapedCity := url.QueryEscape(city.Name)
	url := fmt.Sprintf(
		"https://api.weatherapi.com/v1/current.json?key=%s&q=%s",
		weatherAPIKey, escapedCity,
	)

	var resp *http.Response

	operation := func() error {
		r, err := http.Get(url)
		if err != nil {
			return err
		}

		if r.StatusCode == http.StatusTooManyRequests {
			r.Body.Close()
			return backoff.Permanent(fmt.Errorf("rate limit exceeded"))
		}

		if r.StatusCode >= 400 {
			body, _ := io.ReadAll(r.Body)
			r.Body.Close()
			return fmt.Errorf("HTTP %d: %s", r.StatusCode, string(body))
		}

		resp = r
		return nil
	}

	if err := backoff.Retry(operation, backoff.NewExponentialBackOff()); err != nil {
		return err
	}

	defer resp.Body.Close()

	// JSON → структура
	var result struct {
		Current struct {
			TempC     float64 `json:"temp_c"`
			Humidity  int     `json:"humidity"`
			WindKph   float64 `json:"wind_kph"`
			Condition struct {
				Text string `json:"text"`
			} `json:"condition"`
		} `json:"current"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	data.CityID = city.ID
	data.Source = "WeatherAPI"
	data.Temperature = result.Current.TempC
	data.Humidity = result.Current.Humidity
	data.WindSpeed = math.Round((result.Current.WindKph/3.6)*100) / 100
	data.CreatedAt = timeNow

	return nil
}

func (w *WeatherClient) saveWeatherData(tx *bun.Tx, data *model.WeatherData) {
	_, err := tx.NewInsert().Model(data).Exec(context.Background())
	if err != nil {
		tx.Rollback()
		log.Errorf("Failed to save data: %v", err)
	}
}

func (w *WeatherClient) saveAggregatedWeatherData(tx *bun.Tx, data *model.AggregatedWeatherData) {
	_, err := tx.NewInsert().Model(data).Exec(context.Background())
	if err != nil {
		tx.Rollback()
		log.Errorf("Failed to save data: %v", err)
	}
}
