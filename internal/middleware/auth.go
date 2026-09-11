package middleware

import (
	stderrors "errors"
	"strings"

	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/security"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(sessionManager *security.LoginSessionManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.Unauthorized)))
		}

		// Check if bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidToken)))
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidToken)))
		}

		// Validate token
		claims, err := sessionManager.ValidateAccess(tokenString)
		if err != nil {
			if stderrors.Is(err, security.ErrSessionStoreUnavailable) {
				appErr := errors.New(errors.AuthServiceUnavailable)
				return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidToken)))
		}

		// Set user context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("token_id", claims.ID)
		c.Locals("session_id", claims.SessionID)

		return c.Next()
	}
}
