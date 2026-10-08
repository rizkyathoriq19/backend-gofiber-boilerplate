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
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "boilerplate-be/docs"
	"boilerplate-be/internal/application"
	"boilerplate-be/internal/config"
	"boilerplate-be/internal/database"
	"boilerplate-be/internal/delivery/websocket"
	"boilerplate-be/internal/middleware"
	"boilerplate-be/internal/module/auth"
	"boilerplate-be/internal/module/rbac"

	"boilerplate-be/internal/shared/security"
	"boilerplate-be/internal/shared/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	err := run(ctx)
	stop()
	if err != nil {
		log.Printf("Server failed: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize config
	cfg := config.New()

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect to Redis: %w", err)
	}
	defer redisClient.Close()

	rateLimitStorage, err := middleware.NewRateLimitStorage(cfg)
	if err != nil {
		return fmt.Errorf("connect rate-limit storage: %w", err)
	}
	defer rateLimitStorage.Close()

	// Initialize cache
	cacheHelper := utils.NewCacheHelper(redisClient, cfg.Redis.DefaultTTL)

	// Initialize JWT manager
	jwtManager := security.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry, cfg.JWT.RefreshExpiry)
	sessionManager := security.NewLoginSessionManager(jwtManager, security.NewRedisSessionStore(redisClient))

	// ==================== Initialize Repositories ====================
	authRepo := auth.NewAuthRepository(db, cacheHelper)
	rbacRepo := rbac.NewRBACRepository(db)

	// ==================== Initialize Use Cases ====================
	authUseCase := auth.NewAuthUseCase(authRepo, sessionManager)
	rbacUseCase := rbac.NewRBACUseCase(rbacRepo)

	// ==================== Initialize Handlers ====================
	authHandler := auth.NewAuthHandler(authUseCase, cfg.JWT.Expiry)
	rbacHandler := rbac.NewRBACHandler(rbacUseCase)

	// ==================== Initialize WebSocket ====================
	wsHub := websocket.NewHub()
	defer wsHub.Shutdown()
	go wsHub.Run()

	app := application.New(cfg, application.Dependencies{
		Auth: authHandler, RBAC: rbacHandler, Authorization: rbacUseCase,
		Sessions: sessionManager, RateLimitStorage: rateLimitStorage, WebSocket: wsHub,
	}, application.Options{Docs: true, Presentation: true})

	return serve(ctx, app, ":"+cfg.App.Port)
}

func serve(ctx context.Context, app *fiber.App, address string) error {
	if ctx.Err() != nil {
		return nil
	}
	listenDone := make(chan error, 1)
	go func() { listenDone <- app.Listen(address) }()
	select {
	case err := <-listenDone:
		return err
	case <-ctx.Done():
	}

	// Cancellation can arrive before Fiber registers its listener. Retry shutdown
	// until Listen returns. Keep dependencies alive while active requests drain.
	// ponytail: drain without a deadline; bounded shutdown needs cooperative handler cancellation.
	retry := time.NewTicker(10 * time.Millisecond)
	defer retry.Stop()
	for {
		shutdownErr := app.Shutdown()
		select {
		case listenErr := <-listenDone:
			return errors.Join(listenErr, shutdownErr)

		case <-retry.C:
		}
	}
}
