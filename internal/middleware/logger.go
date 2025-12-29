package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// LoggerMiddleware returns a Fiber logger middleware configured based on environment
func LoggerMiddleware(env string) fiber.Handler {
	// Production: JSON format for structured logging
	if env == "production" {
		return logger.New(logger.Config{
			Format:     `{"time":"${time}","status":${status},"latency":"${latency}","ip":"${ip}","method":"${method}","path":"${path}","error":"${error}"}` + "\n",
			TimeFormat: "2006-01-02T15:04:05Z07:00",
			TimeZone:   "Local",
			Output:     os.Stdout,
		})
	}

	// Development: Human-readable format with colors
	return logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
		Output:     os.Stdout,
	})
}
