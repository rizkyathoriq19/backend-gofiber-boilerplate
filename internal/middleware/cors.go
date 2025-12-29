package middleware

import (
	"strings"

	"boilerplate-be/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsMiddleware(cfg *config.Config) fiber.Handler {
	origins := strings.Join(cfg.CORS.AllowedOrigins, ",")

	allowCredentials := true
	if origins == "*" {
		allowCredentials = false
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     strings.Join(cfg.CORS.AllowedMethods, ","),
		AllowHeaders:     strings.Join(cfg.CORS.AllowedHeaders, ","),
		AllowCredentials: allowCredentials,
	})
}
