package middleware

import (
	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware(logger *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		err := c.Next()
		
		duration := time.Since(start)
		
		logger.WithFields(logrus.Fields{
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     c.Response().StatusCode(),
			"duration":   duration.Milliseconds(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		}).Info("HTTP Request")
		
		return err
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	if appErr, ok := errors.IsAppError(err); ok {
		return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
	}
	
	// Default error response
	return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(errors.ErrServerError))
}