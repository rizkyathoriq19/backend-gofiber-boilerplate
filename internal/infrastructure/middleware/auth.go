package middleware

import (
	"context"
	"strings"
	"time"

	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/redis"
	"boilerplate-be/internal/infrastructure/token"
	"boilerplate-be/internal/response"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtManager *token.JWTManager, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.ErrUnauthorized))
		}

		// Check if bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.ErrInvalidToken))
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.ErrInvalidToken))
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.ErrInvalidToken))
		}

		// Check if token is access token
		if claims.TokenType != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.ErrInvalidToken))
		}

		// Check if token is blacklisted
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		isBlacklisted, err := jwtManager.IsTokenBlacklisted(ctx, redisClient, claims.ID)
		if err != nil {
			// If Redis is down, log the error but allow the request to proceed
			// In production, you might want to handle this differently
		} else if isBlacklisted {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.ErrInvalidToken))
		}

		// Set user context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("token_id", claims.ID)

		return c.Next()
	}
}