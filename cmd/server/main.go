// Package main is the entry point for the Go Fiber Boilerplate API
//
// @title           Go Fiber Boilerplate API
// @version         1.0
// @description     A production-ready Go Fiber boilerplate with RBAC, Redis caching, and PostgreSQL.
//
// @contact.name   API Support
// @contact.email  support@example.com
//
// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT
//
// @host      localhost:8000
// @BasePath  /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Format: Bearer {token}. Get token from /auth/login endpoint.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "boilerplate-be/docs"
	"boilerplate-be/internal/config"
	"boilerplate-be/internal/database"
	"boilerplate-be/internal/database/redis"
	"boilerplate-be/internal/middleware"
	"boilerplate-be/internal/module/auth"
	"boilerplate-be/internal/module/rbac"
	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"
	"boilerplate-be/internal/pkg/security"
	"boilerplate-be/internal/pkg/utils"
	"boilerplate-be/web"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize config
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.New(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize token manager
	tokenManager := security.NewTokenManager(redisClient)

	// Initialize cache
	cacheHelper := utils.NewCacheHelper(redisClient, cfg.Redis.DefaultTTL)

	// Initialize JWT manager
	jwtManager := security.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// ==================== Initialize Repositories ====================
	authRepo := auth.NewAuthRepository(db, cacheHelper)
	rbacRepo := rbac.NewRBACRepository(db, cacheHelper)

	// ==================== Initialize Use Cases ====================
	authUseCase := auth.NewAuthUseCase(authRepo, jwtManager, tokenManager)
	rbacUseCase := rbac.NewRBACUseCase(rbacRepo)

	// ==================== Initialize Handlers ====================
	authHandler := auth.NewAuthHandler(authUseCase)
	rbacHandler := rbac.NewRBACHandler(rbacUseCase)

	// Initialize Fiber app with optimized config
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
	app.Use(middleware.RateLimitMiddleware(cfg))

	// Swagger UI
	app.Get("/swagger/*", swagger.New(swagger.Config{
		DeepLinking: true,
	}))

	// Serve static docs files (architecture diagrams, etc)
	app.Static("/docs", "./docs")

	// Health check endpoints
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// Routes
	api := app.Group("/api/v1")

	// ==================== Public Routes ====================
	// Auth routes (public)
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.RefreshToken)

	// ==================== Protected Routes (Authenticated Users) ====================
	// Auth routes (protected)
	authProtected := authGroup.Group("", middleware.AuthMiddleware(jwtManager, redisClient))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/profile", authHandler.Profile)
	authProtected.Put("/profile", authHandler.UpdateProfile)
	authProtected.Get("/my-roles", rbacHandler.GetMyRoles)
	authProtected.Get("/my-permissions", rbacHandler.GetMyPermissions)

	// ==================== Super Admin Routes ====================
	// Super admin routes (requires super_admin role)
	superAdmin := api.Group("/super-admin",
		middleware.AuthMiddleware(jwtManager, redisClient),
		middleware.IsSuperAdmin(rbacUseCase),
	)

	// User role management
	superAdmin.Get("/users/:userId/roles", rbacHandler.GetUserRoles)
	superAdmin.Post("/users/:userId/roles", rbacHandler.AssignRoleToUser)
	superAdmin.Delete("/users/:userId/roles/:roleId", rbacHandler.RemoveRoleFromUser)

	// Role management
	superAdmin.Get("/roles", rbacHandler.GetRoles)
	superAdmin.Get("/roles/:id", rbacHandler.GetRole)
	superAdmin.Post("/roles", rbacHandler.CreateRole)
	superAdmin.Put("/roles/:id", rbacHandler.UpdateRole)
	superAdmin.Delete("/roles/:id", rbacHandler.DeleteRole)

	// Permission management
	superAdmin.Get("/permissions", rbacHandler.GetPermissions)
	superAdmin.Get("/roles/:id/permissions", rbacHandler.GetRolePermissions)
	superAdmin.Post("/roles/:id/permissions", rbacHandler.AssignPermissionToRole)
	superAdmin.Delete("/roles/:id/permissions/:permissionId", rbacHandler.RemovePermissionFromRole)


	// Health check - HTML UI
	api.Get("/health", func(c *fiber.Ctx) error {
		html, err := web.RenderHealth(cfg.App.Name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error rendering page")
		}
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// Root path handler - Welcome UI
	app.Get("/", func(c *fiber.Ctx) error {
		html, err := web.RenderIndex(cfg.App.Name)
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

	// 404 Not Found handler - HTML UI
	app.Use(func(c *fiber.Ctx) error {
		// Check Accept header - return JSON for API clients
		acceptHeader := c.Get("Accept")
		if acceptHeader == "application/json" {
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

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.App.Port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/", cfg.App.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
