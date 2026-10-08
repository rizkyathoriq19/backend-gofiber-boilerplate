// Package application assembles production HTTP routes and middleware from caller-owned dependencies.
package application

import (
	_ "boilerplate-be/docs"
	"boilerplate-be/internal/config"
	"boilerplate-be/internal/delivery/websocket"
	"boilerplate-be/internal/middleware"
	"boilerplate-be/internal/module/auth"
	"boilerplate-be/internal/module/rbac"
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/security"
	"boilerplate-be/web"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// Dependencies are created and closed by the caller. A nil WebSocket omits its routes.
type Dependencies struct {
	Auth             *auth.AuthHandler
	RBAC             *rbac.RBACHandler
	Authorization    rbac.RBACUseCase
	Sessions         *security.LoginSessionManager
	RateLimitStorage fiber.Storage
	WebSocket        *websocket.Hub
}

// Options explicitly compose optional route groups; they are not environment flags.
type Options struct {
	Docs         bool
	Presentation bool
}

// New neither acquires infrastructure nor starts listeners or background workers.
// The caller must supply rate-limit storage and keep dependencies alive until shutdown.
func New(cfg *config.Config, deps Dependencies, options Options) *fiber.App {
	if deps.RateLimitStorage == nil {
		panic("application: rate-limit storage is required")
	}
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: middleware.ErrorHandler,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		Prefork:      cfg.App.Prefork,
		// Performance optimizations
		ReduceMemoryUsage:     true,
		DisableStartupMessage: cfg.App.Env == "production",
		ReadBufferSize:        4096,
		WriteBufferSize:       4096,
	})

	// Add middleware (order matters!)
	app.Use(recover.New())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.LoggerMiddleware(cfg.App.Env))
	app.Use(compress.New())
	app.Use(etag.New())
	app.Use(middleware.CorsMiddleware(cfg))
	app.Use(middleware.HelmetMiddleware())
	app.Use(middleware.RateLimitMiddleware(cfg, deps.RateLimitStorage))

	if options.Docs {
		app.Get("/swagger/*", swagger.New(swagger.Config{DeepLinking: true}))
		app.Static("/docs", "./docs")
	}
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendString("pong") })
	if deps.WebSocket != nil {
		websocket.RegisterRoutes(app, deps.WebSocket)
	}
	// Routes
	api := app.Group("/api/v1")

	// ==================== Public Routes ====================
	// Auth routes (public)
	authGroup := api.Group("/auth")
	authGroup.Post("/register", deps.Auth.Register)
	authGroup.Post("/login", deps.Auth.Login)
	authGroup.Post("/refresh", deps.Auth.RefreshToken)

	// ==================== Protected Routes (Authenticated Users) ====================
	// Auth routes (protected)
	authProtected := authGroup.Group("", middleware.AuthMiddleware(deps.Sessions))
	authProtected.Post("/logout", deps.Auth.Logout)
	authProtected.Get("/profile", deps.Auth.Profile)
	authProtected.Put("/profile", deps.Auth.UpdateProfile)
	authProtected.Get("/my-roles", deps.RBAC.GetMyRoles)
	authProtected.Get("/my-permissions", deps.RBAC.GetMyPermissions)

	// ==================== Super Admin Routes ====================
	// Super admin routes (requires super_admin role)
	superAdmin := api.Group("/super-admin",
		middleware.AuthMiddleware(deps.Sessions),
		middleware.IsSuperAdmin(deps.Authorization),
	)

	// User role management
	superAdmin.Get("/users/:userId/roles", deps.RBAC.GetUserRoles)
	superAdmin.Post("/users/:userId/roles", deps.RBAC.AssignRoleToUser)
	superAdmin.Delete("/users/:userId/roles/:roleId", deps.RBAC.RemoveRoleFromUser)

	// Role management
	superAdmin.Get("/roles", deps.RBAC.GetRoles)
	superAdmin.Get("/roles/:id", deps.RBAC.GetRole)
	superAdmin.Post("/roles", deps.RBAC.CreateRole)
	superAdmin.Put("/roles/:id", deps.RBAC.UpdateRole)
	superAdmin.Delete("/roles/:id", deps.RBAC.DeleteRole)

	// Permission management
	superAdmin.Get("/permissions", deps.RBAC.GetPermissions)
	superAdmin.Get("/roles/:id/permissions", deps.RBAC.GetRolePermissions)
	superAdmin.Post("/roles/:id/permissions", deps.RBAC.AssignPermissionToRole)
	superAdmin.Delete("/roles/:id/permissions/:permissionId", deps.RBAC.RemovePermissionFromRole)

	if options.Presentation {
		registerPresentation(app, cfg.App.Name)
	}
	// 404 Not Found handler - HTML UI
	app.Use(func(c *fiber.Ctx) error {
		// Check Accept header - return JSON for API clients
		acceptHeader := c.Get("Accept")
		if !options.Presentation || acceptHeader == "application/json" {
			return c.Status(fiber.StatusNotFound).JSON(
				response.CreateErrorResponse(c, errors.New(errors.ResourceNotFound)),
			)
		}

		// Return HTML for browser routes
		var html string
		var err error

		// Use not_found template for API routes, 404 for other routes
		if len(c.Path()) > 4 && c.Path()[:4] == "/api" {
			html, err = web.RenderNotFound()
		} else {
			html, err = web.Render404()
		}

		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Page Not Found")
		}
		c.Set("Content-Type", "text/html")
		return c.Status(fiber.StatusNotFound).SendString(html)
	})

	return app
}

func registerPresentation(app *fiber.App, appName string) {
	// Health check - HTML UI
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		html, err := web.RenderHealth(appName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error rendering page")
		}
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// Root path handler - Welcome UI
	app.Get("/", func(c *fiber.Ctx) error {
		html, err := web.RenderIndex(appName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error rendering page")
		}
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// Resource not found page
	app.Get("/not-found", func(c *fiber.Ctx) error {
		html, err := web.RenderNotFound()
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Resource Not Found")
		}
		c.Set("Content-Type", "text/html")
		return c.Status(fiber.StatusNotFound).SendString(html)
	})

}
