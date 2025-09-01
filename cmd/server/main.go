package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"boilerplate-be/internal/infrastructure/config"
	"boilerplate-be/internal/infrastructure/database"
	"boilerplate-be/internal/infrastructure/helper"
	"boilerplate-be/internal/infrastructure/logger"
	"boilerplate-be/internal/infrastructure/middleware"
	"boilerplate-be/internal/infrastructure/redis"
	"boilerplate-be/internal/infrastructure/token"
	"boilerplate-be/internal/module/auth/handler"
	"boilerplate-be/internal/module/auth/repository"
	"boilerplate-be/internal/module/auth/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize config
	cfg := config.New()

	// Initialize logger
	logger := logger.New(cfg.App.Env)

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.New(cfg.Redis)
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize token manager
	tokenManager := helper.NewTokenManager(redisClient)

	// Initialize cache
	cacheHelper := helper.NewCacheHelper(redisClient, cfg.Redis.DefaultTTL)

	// Initialize JWT manager
	jwtManager := token.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Initialize repositories
	authRepo := repository.NewAuthRepository(db, cacheHelper)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(authRepo, jwtManager, tokenManager)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: middleware.ErrorHandler,
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(middleware.LoggerMiddleware(logger))
	app.Use(middleware.CorsMiddleware(cfg))
	app.Use(middleware.HelmetMiddleware())
	app.Use(middleware.RateLimitMiddleware(redisClient, cfg))

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	
	// Routes
	api := app.Group("/api/v1")
	
	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", middleware.AuthMiddleware(jwtManager, redisClient), authHandler.Logout)
	auth.Get("/profile", middleware.AuthMiddleware(jwtManager, redisClient), authHandler.Profile)
	auth.Put("/profile", middleware.AuthMiddleware(jwtManager, redisClient), authHandler.UpdateProfile)

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"message": "Server is running",
		})
	})

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	logger.Infof("Server started on port %s", cfg.App.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}