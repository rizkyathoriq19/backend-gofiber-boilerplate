package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"boilerplate-be/internal/infrastructure/config"
	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/redis"
	"boilerplate-be/internal/response"

	"github.com/gofiber/fiber/v2"
)

func RateLimitMiddleware(redisClient *redis.Client, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client IP
		ip := c.IP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Check current count
		currentStr, err := redisClient.GetValue(ctx, key)
		if err != nil && err.Error() != "redis: nil" {
			// If Redis is down, allow request to proceed
			return c.Next()
		}

		var current int64 = 0
		if currentStr != "" {
			current, _ = strconv.ParseInt(currentStr, 10, 64)
		}

		// Check if limit exceeded
		if current >= int64(cfg.RateLimit.Max) {
			rateLimitError := errors.New(errors.RateLimitExceeded, "en")
			return c.Status(rateLimitError.StatusCode).JSON(response.ErrorResponse(rateLimitError))
		}

		// Increment counter
		newCount, err := redisClient.Incr(ctx, key)
		if err != nil {
			// If Redis is down, allow request to proceed
			return c.Next()
		}

		// Set expiration only for the first request
		if newCount == 1 {
			redisClient.Expire(ctx, key, cfg.RateLimit.Window)
		}

		// Add rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimit.Max))
		c.Set("X-RateLimit-Remaining", strconv.FormatInt(int64(cfg.RateLimit.Max)-newCount, 10))
		
		// Get TTL for reset time
		ttl, err := redisClient.TTL(ctx, key)
		if err == nil && ttl > 0 {
			resetTime := time.Now().Add(ttl).Unix()
			c.Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		}

		return c.Next()
	}
}

// Alternative simpler rate limiter for specific endpoints
func EndpointRateLimitMiddleware(redisClient *redis.Client, maxRequests int, window time.Duration, keyPrefix string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("%s:%s", keyPrefix, ip)
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get current count
		currentStr, err := redisClient.GetValue(ctx, key)
		var current int64 = 0
		if err == nil && currentStr != "" {
			current, _ = strconv.ParseInt(currentStr, 10, 64)
		}

		// Check limit
		if current >= int64(maxRequests) {
			rateLimitError := errors.New(errors.RateLimitExceeded, "en")
			return c.Status(rateLimitError.StatusCode).JSON(response.ErrorResponse(rateLimitError))
		}

		// Increment
		newCount, err := redisClient.Incr(ctx, key)
		if err == nil && newCount == 1 {
			redisClient.Expire(ctx, key, window)
		}

		return c.Next()
	}
}