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
// @host      localhost:3002
// @BasePath  /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Format: Bearer {token}. Get token from /auth/login endpoint.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"boilerplate-be/docs"
	"boilerplate-be/internal/config"
	"boilerplate-be/internal/database"
	"boilerplate-be/internal/delivery/websocket"
	"boilerplate-be/internal/middleware"
	"boilerplate-be/internal/module/alert"
	"boilerplate-be/internal/module/auth"
	"boilerplate-be/internal/module/device"
	"boilerplate-be/internal/module/message"
	"boilerplate-be/internal/module/patient"
	"boilerplate-be/internal/module/rbac"
	"boilerplate-be/internal/module/room"
	"boilerplate-be/internal/module/staff"
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/security"
	"boilerplate-be/internal/shared/utils"
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
	redisClient, err := database.NewRedis(cfg.Redis)
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
	roomRepo := room.NewRoomRepository(db, cacheHelper)
	deviceRepo := device.NewDeviceRepository(db, cacheHelper)
	staffRepo := staff.NewStaffRepository(db, cacheHelper)
	patientRepo := patient.NewPatientRepository(db, cacheHelper)
	alertRepo := alert.NewAlertRepository(db, cacheHelper)
	messageRepo := message.NewMessageRepository(db, cacheHelper)

	// ==================== Initialize Use Cases ====================
	authUseCase := auth.NewAuthUseCase(authRepo, jwtManager, tokenManager)
	rbacUseCase := rbac.NewRBACUseCase(rbacRepo)
	roomUseCase := room.NewRoomUseCase(roomRepo)
	deviceUseCase := device.NewDeviceUseCase(deviceRepo)
	staffUseCase := staff.NewStaffUseCase(staffRepo)
	patientUseCase := patient.NewPatientUseCase(patientRepo)
	alertUseCase := alert.NewAlertUseCase(alertRepo, staffRepo, patientRepo, deviceRepo)
	messageUseCase := message.NewMessageUseCase(messageRepo)

	// ==================== Initialize Handlers ====================
	authHandler := auth.NewAuthHandler(authUseCase)
	rbacHandler := rbac.NewRBACHandler(rbacUseCase)
	roomHandler := room.NewRoomHandler(roomUseCase)
	deviceHandler := device.NewDeviceHandler(deviceUseCase)
	staffHandler := staff.NewStaffHandler(staffUseCase)
	patientHandler := patient.NewPatientHandler(patientUseCase)
	alertHandler := alert.NewAlertHandler(alertUseCase)
	messageHandler := message.NewMessageHandler(messageUseCase)

	// ==================== Initialize WebSocket ====================
	wsHub := websocket.NewHub()
	go wsHub.Run()

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
		// Trust proxy headers (for devtunnel, ngrok, cloudflare, etc.)
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1"},
		ProxyHeader:             fiber.HeaderXForwardedFor,
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

	// Configure Swagger docs with dynamic host
	hosts := cfg.Swagger.Hosts
	schemes := cfg.Swagger.Schemes
	basePath := cfg.Swagger.BasePath

	// Set default host
	if len(hosts) > 0 {
		docs.SwaggerInfo.Host = hosts[0]
	}
	docs.SwaggerInfo.BasePath = basePath
	docs.SwaggerInfo.Schemes = schemes

	// Build URLs config for swagger (multiple servers dropdown)
	var urlsConfig string
	for i, host := range hosts {
		name := host
		if i == 0 {
			name = "Local - " + host
		} else {
			name = "Remote - " + host
		}
		if i > 0 {
			urlsConfig += ","
		}
		urlsConfig += fmt.Sprintf(`{url: "/swagger/doc.json?host=%d", name: "%s"}`, i, name)
	}

	// Serve dynamic swagger doc.json based on host parameter
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		hostIdx := c.QueryInt("host", 0)
		if hostIdx >= 0 && hostIdx < len(hosts) {
			docs.SwaggerInfo.Host = hosts[hostIdx]
		}
		c.Set("Content-Type", "application/json")
		return c.SendString(docs.SwaggerInfo.ReadDoc())
	})

	// Custom Swagger UI with server selector dropdown
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusMovedPermanently)
	})

	app.Get("/swagger/*", func(c *fiber.Ctx) error {
		path := c.Params("*")
		if path == "" || path == "index.html" {
			c.Set("Content-Type", "text/html")
			html, err := web.RenderSwagger(cfg.App.Name, urlsConfig, "Local - "+hosts[0])
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Error rendering swagger page")
			}
			return c.SendString(html)
		}
		// For other swagger assets, use default handler
		return swagger.New(swagger.Config{
			DeepLinking: true,
		})(c)
	})

	// Serve static docs files (architecture diagrams, etc)
	app.Static("/docs", "./docs")

	// Health check endpoints
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// ==================== WebSocket Routes ====================
	websocket.RegisterRoutes(app, wsHub)

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
	
	// Batch permission management
	superAdmin.Post("/roles/:id/permissions/batch", rbacHandler.BatchAssignPermissionsToRole)
	superAdmin.Delete("/roles/:id/permissions/batch", rbacHandler.BatchRemovePermissionsFromRole)
	superAdmin.Post("/roles/permissions/batch", rbacHandler.BatchGetRolePermissions)

	// ==================== MEDIPROMPT Protected Routes ====================
	// All MEDIPROMPT routes require authentication
	protected := api.Group("", middleware.AuthMiddleware(jwtManager, redisClient))

	// Room routes (Admin/Manager only for write operations)
	roomsGroup := protected.Group("/rooms")
	roomsGroup.Get("", roomHandler.GetRooms)
	roomsGroup.Get("/:id", roomHandler.GetRoom)
	roomsGroup.Post("", roomHandler.CreateRoom)       // TODO: Add admin/manager middleware
	roomsGroup.Put("/:id", roomHandler.UpdateRoom)    // TODO: Add admin/manager middleware
	roomsGroup.Delete("/:id", roomHandler.DeleteRoom) // TODO: Add admin/manager middleware

	// Device routes
	devicesGroup := protected.Group("/devices")
	devicesGroup.Get("", deviceHandler.GetDevices)
	devicesGroup.Get("/:id", deviceHandler.GetDevice)
	devicesGroup.Post("", deviceHandler.RegisterDevice) // Returns API key
	devicesGroup.Put("/:id", deviceHandler.UpdateDevice)
	devicesGroup.Put("/:id/status", deviceHandler.UpdateDeviceStatus)
	devicesGroup.Delete("/:id", deviceHandler.DeleteDevice)

	// Device heartbeat (uses X-API-Key, no JWT)
	api.Post("/devices/heartbeat", deviceHandler.Heartbeat)

	// Staff routes
	staffGroup := protected.Group("/staff")
	staffGroup.Get("", staffHandler.GetAllStaff)
	staffGroup.Get("/on-duty", staffHandler.GetOnDutyStaff)
	staffGroup.Get("/:id", staffHandler.GetStaff)
	staffGroup.Post("", staffHandler.CreateStaff)
	staffGroup.Put("/:id", staffHandler.UpdateStaff)
	staffGroup.Put("/:id/shift", staffHandler.UpdateShift)
	staffGroup.Post("/:id/toggle-duty", staffHandler.ToggleOnDuty)
	staffGroup.Delete("/:id", staffHandler.DeleteStaff)
	staffGroup.Get("/:id/rooms", staffHandler.GetRoomAssignments)
	staffGroup.Post("/:id/rooms", staffHandler.AssignToRoom)
	staffGroup.Delete("/:id/rooms/:roomId", staffHandler.RemoveFromRoom)

	// Patient routes
	patientsGroup := protected.Group("/patients")
	patientsGroup.Get("", patientHandler.GetPatients)
	patientsGroup.Get("/:id", patientHandler.GetPatient)
	patientsGroup.Post("", patientHandler.AdmitPatient)
	patientsGroup.Put("/:id", patientHandler.UpdatePatient)
	patientsGroup.Put("/:id/condition", patientHandler.UpdateConditionLevel)
	patientsGroup.Post("/:id/discharge", patientHandler.DischargePatient)
	patientsGroup.Delete("/:id", patientHandler.DeletePatient)

	// Alert routes
	alertsGroup := protected.Group("/alerts")
	alertsGroup.Get("", alertHandler.GetAlerts)
	alertsGroup.Get("/active", alertHandler.GetActiveAlerts)
	alertsGroup.Get("/:id", alertHandler.GetAlert)
	alertsGroup.Post("", alertHandler.CreateAlert)
	alertsGroup.Post("/:id/acknowledge", alertHandler.AcknowledgeAlert)
	alertsGroup.Post("/:id/resolve", alertHandler.ResolveAlert)
	alertsGroup.Get("/:id/history", alertHandler.GetAlertHistory)

	// Message routes
	messagesGroup := protected.Group("/messages")
	messagesGroup.Get("", messageHandler.GetMessages)
	messagesGroup.Get("/my", messageHandler.GetMyMessages)
	messagesGroup.Get("/unread-count", messageHandler.GetUnreadCount)
	messagesGroup.Get("/:id", messageHandler.GetMessage)
	messagesGroup.Post("", messageHandler.SendMessage)
	messagesGroup.Put("/:id/read", messageHandler.MarkAsRead)
	messagesGroup.Put("/read-all", messageHandler.MarkAllAsRead)
	messagesGroup.Delete("/:id", messageHandler.DeleteMessage)


	// Health check - HTML UI
	app.Get("/health", func(c *fiber.Ctx) error {
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
