package middleware

import (
	"boilerplate-be/internal/module/rbac"
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"

	"github.com/gofiber/fiber/v2"
)

// RequireRole creates a middleware that checks if the user has any of the required roles
func RequireRole(rbacUseCase rbac.RBACUseCase, roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			appErr := errors.New(errors.Unauthorized)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		hasRole, err := rbacUseCase.CheckUserRole(userID, roles...)
		if err != nil {
			appErr := errors.New(errors.InternalServerError)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		if !hasRole {
			appErr := errors.New(errors.Forbidden)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		return c.Next()
	}
}

// RequirePermission creates a middleware that checks if the user has any of the required permissions
func RequirePermission(rbacUseCase rbac.RBACUseCase, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			appErr := errors.New(errors.Unauthorized)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		hasPermission, err := rbacUseCase.CheckUserPermission(userID, permissions...)
		if err != nil {
			appErr := errors.New(errors.InternalServerError)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		if !hasPermission {
			appErr := errors.New(errors.Forbidden)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		return c.Next()
	}
}

// RequireAnyRole creates a middleware that checks if the user has at least one of the required roles
// This is an alias for RequireRole for semantic clarity
func RequireAnyRole(rbacUseCase rbac.RBACUseCase, roles ...string) fiber.Handler {
	return RequireRole(rbacUseCase, roles...)
}

// RequireAllRoles creates a middleware that checks if the user has all of the required roles
func RequireAllRoles(rbacUseCase rbac.RBACUseCase, roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			appErr := errors.New(errors.Unauthorized)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		for _, role := range roles {
			hasRole, err := rbacUseCase.CheckUserRole(userID, role)
			if err != nil {
				appErr := errors.New(errors.InternalServerError)
				return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
			}
			if !hasRole {
				appErr := errors.New(errors.Forbidden)
				return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
			}
		}

		return c.Next()
	}
}

// RequireAllPermissions creates a middleware that checks if the user has all of the required permissions
func RequireAllPermissions(rbacUseCase rbac.RBACUseCase, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			appErr := errors.New(errors.Unauthorized)
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}

		for _, permission := range permissions {
			hasPermission, err := rbacUseCase.CheckUserPermission(userID, permission)
			if err != nil {
				appErr := errors.New(errors.InternalServerError)
				return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
			}
			if !hasPermission {
				appErr := errors.New(errors.Forbidden)
				return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
			}
		}

		return c.Next()
	}
}

// IsSuperAdmin is a convenience middleware for super admin only routes
func IsSuperAdmin(rbacUseCase rbac.RBACUseCase) fiber.Handler {
	return RequireRole(rbacUseCase, "super_admin")
}
