package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/helmet/v2"
)

func HelmetMiddleware() fiber.Handler {
	return helmet.New()
}