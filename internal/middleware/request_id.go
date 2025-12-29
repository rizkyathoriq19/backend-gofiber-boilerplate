package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request for tracing
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists (from load balancer or gateway)
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in response header
		c.Set("X-Request-ID", requestID)

		// Store in locals for access in handlers and logging
		c.Locals("request_id", requestID)

		return c.Next()
	}
}
