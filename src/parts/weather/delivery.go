package weather

import (
	"github.com/gofiber/fiber/v2"
)

// Controller represent controllers
type Controller interface {
	HealthCheck(c *fiber.Ctx) error
	GetCurrent(c *fiber.Ctx) error
	GetForecast(c *fiber.Ctx) error
}
