package middleware

import (
	"boilerplate-be/internal/config"
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/redis/v3"
	goredis "github.com/redis/go-redis/v9"
)

// NewRateLimitStorage creates caller-owned Redis storage. Close it after HTTP shutdown.
func NewRateLimitStorage(cfg *config.Config) (*redis.Storage, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, strconv.Itoa(parsePort(cfg.Redis.Port))),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect rate-limit storage: %w", err)
	}
	return redis.NewFromConnection(client), nil
}

// RateLimitMiddleware uses supplied storage without acquiring or owning infrastructure.
func RateLimitMiddleware(cfg *config.Config, storage fiber.Storage) fiber.Handler {
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
func EndpointRateLimitMiddleware(cfg *config.Config, storage fiber.Storage, maxRequests int, keyPrefix string) fiber.Handler {
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
