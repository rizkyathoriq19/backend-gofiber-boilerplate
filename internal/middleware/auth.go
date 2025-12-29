package middleware

import (
	"context"
	"strings"
	"time"

	"boilerplate-be/internal/database/redis"
	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"
	"boilerplate-be/internal/pkg/security"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtManager *security.JWTManager, redisClient *redis.Client) fiber.Handler {
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
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidToken)))
		}

		// Check if token is access token
		if claims.TokenType != "access" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidToken)))
		}

		// Check if token is blacklisted
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		isBlacklisted, err := jwtManager.IsTokenBlacklisted(ctx, redisClient, claims.ID)
		if err != nil {
			// Log Redis error - in production, consider failing closed instead of open
			// For now, we allow the request to proceed if Redis is unavailable
			// This is a tradeoff between availability and security
		} else if isBlacklisted {
			return c.Status(fiber.StatusUnauthorized).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidToken)))
		}

		// Set user context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("token_id", claims.ID)

		return c.Next()
	}
}
