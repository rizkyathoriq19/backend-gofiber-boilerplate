package middleware

import (
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is the centralized error handler for Fiber
func ErrorHandler(c *fiber.Ctx, err error) error {
	if appErr, ok := errors.IsAppError(err); ok {
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	// Default error response
	return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
}
