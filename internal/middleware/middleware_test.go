package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(RequestIDMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		requestID := c.Locals("request_id")
		return c.JSON(fiber.Map{"request_id": requestID})
	})

	t.Run("generates new request ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}

		requestID := resp.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Error("X-Request-ID header should be set")
		}
	})

	t.Run("uses existing request ID when provided", func(t *testing.T) {
		existingID := "existing-request-id-123"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("failed to execute request: %v", err)
		}

		responseID := resp.Header.Get("X-Request-ID")
		if responseID != existingID {
			t.Errorf("expected request ID '%s', got '%s'", existingID, responseID)
		}
	})
}

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{"production environment", "production"},
		{"development environment", "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(LoggerMiddleware(tt.env))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}

			if resp.StatusCode != fiber.StatusOK {
				t.Errorf("expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestHelmetMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	// Check security headers are set
	headers := []string{
		"X-XSS-Protection",
		"X-Content-Type-Options",
		"X-Frame-Options",
	}

	for _, header := range headers {
		if resp.Header.Get(header) == "" {
			t.Errorf("expected %s header to be set", header)
		}
	}
}

func TestCorsMiddleware(t *testing.T) {
	// This test requires config, skipping for now
	t.Skip("CORS middleware requires full config setup")
}
