package scheduled_tasks

import (
	"github.com/gofiber/fiber/v2/log"
	crn "github.com/robfig/cron/v3"
	"weather-data-aggregator-service/src/infrastructure/weather"
)

func RunCronJobs(weatherClient *weather.WeatherClient) {
	cronJobRunner := crn.New()

	err := weather.InitWeatherCronJobs(cronJobRunner, weatherClient)
	if err != nil {
		log.Fatalf("Failed to init cron jobs: %s", err)
	}

	cronJobRunner.Start()
}
