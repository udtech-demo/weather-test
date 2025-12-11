package weather

import (
	"fmt"
	crn "github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

func InitWeatherAPI(db *bun.DB) *WeatherClient {
	return createWeatherClient(db)
}

func InitWeatherCronJobs(cronJobRunner *crn.Cron, c *WeatherClient) error {

	if _, err := cronJobRunner.AddFunc("*/15 * * * *", c.fetchWeatherData); err != nil {
		return fmt.Errorf("InitWeatherCronJobs: %s", err)
	}

	return nil
}
