package http

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"weather-data-aggregator-service/src/domain/model"
	"weather-data-aggregator-service/src/parts/weather"
)

type weatherController struct {
	useCase weather.UseCase
}

func NewWeatherController(useCase weather.UseCase) weather.Controller {
	return &weatherController{useCase}
}

// HealthCheck
func (u *weatherController) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "OK"})
}

// GetCurrent returns current aggregated weather for specified city
func (u *weatherController) GetCurrent(c *fiber.Ctx) error {
	var q model.CurrentQuery

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}

	if q.City == "" {
		return fiber.NewError(fiber.StatusBadRequest, "city is required")
	}

	result, err := u.useCase.GetCurrent(c.Context(), q)
	if err != nil {
		return fmt.Errorf("failed to get current weather: %w", err)
	}

	return c.JSON(result)
}

// GetForecast returns aggregated forecast data with validated 'days' parameter
func (u *weatherController) GetForecast(c *fiber.Ctx) error {
	var q model.ForecastQuery

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}

	if q.City == "" {
		return fiber.NewError(fiber.StatusBadRequest, "city is required")
	}

	if q.Days < 1 || q.Days > 7 {
		return fiber.NewError(fiber.StatusBadRequest, "days must be between 1 and 7")
	}

	result, err := u.useCase.GetForecast(c.Context(), q)
	if err != nil {
		return fmt.Errorf("failed to get forecast: %w", err)
	}

	return c.JSON(result)
}
