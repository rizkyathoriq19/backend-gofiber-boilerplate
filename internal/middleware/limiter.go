package middleware

import (
	"boilerplate-be/internal/config"
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
)

// RateLimitMiddleware creates a rate limiter using Fiber's built-in limiter with Redis storage
func RateLimitMiddleware(cfg *config.Config) fiber.Handler {
	// Create Redis storage for rate limiter
	storage := redis.New(redis.Config{
		Host:     cfg.Redis.Host,
		Port:     parsePort(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		Database: cfg.Redis.DB,
	})

	return limiter.New(limiter.Config{
		Max:        cfg.RateLimit.Max,
		Expiration: cfg.RateLimit.Window,
		Storage:    storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			rateLimitError := errors.New(errors.RateLimitExceeded)
			return c.Status(rateLimitError.StatusCode).JSON(
				response.CreateErrorResponse(c, rateLimitError),
			)
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// EndpointRateLimitMiddleware creates a rate limiter for specific endpoints
func EndpointRateLimitMiddleware(cfg *config.Config, maxRequests int, keyPrefix string) fiber.Handler {
	storage := redis.New(redis.Config{
		Host:     cfg.Redis.Host,
		Port:     parsePort(cfg.Redis.Port),
		Password: cfg.Redis.Password,
		Database: cfg.Redis.DB,
	})

	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: cfg.RateLimit.Window,
		Storage:    storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			return keyPrefix + ":" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			rateLimitError := errors.New(errors.RateLimitExceeded)
			return c.Status(rateLimitError.StatusCode).JSON(
				response.CreateErrorResponse(c, rateLimitError),
			)
		},
	})
}

// parsePort converts string port to int
func parsePort(port string) int {
	var p int
	for _, c := range port {
		if c >= '0' && c <= '9' {
			p = p*10 + int(c-'0')
		}
	}
	if p == 0 {
		return 6379 // default Redis port
	}
	return p
}
