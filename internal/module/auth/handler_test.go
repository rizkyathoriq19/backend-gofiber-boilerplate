package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"boilerplate-be/internal/middleware"
	apperrors "boilerplate-be/internal/shared/errors"

	"github.com/gofiber/fiber/v2"
)

func setupTestApp(authHandler *AuthHandler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Post("/register", authHandler.Register)
	app.Post("/login", authHandler.Login)
	return app
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid registration",
			requestBody: map[string]interface{}{
				"email": "newuser@example.com", "password": "password123", "name": "New User",
			},
			expectedStatus: fiber.StatusCreated,
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123", "name": "New User",
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com", "name": "Test User",
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email": "invalid-email", "password": "password123", "name": "Test User",
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{name: "empty body", requestBody: map[string]interface{}{}, expectedStatus: fiber.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp(NewAuthHandler(&mockAuthUseCase{}, time.Hour))
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		loginErr       error
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "valid login",
			requestBody:    map[string]interface{}{"email": "test@example.com", "password": "password123"},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "user not found",
			loginErr:       apperrors.New(apperrors.AccountNotFound),
			requestBody:    map[string]interface{}{"email": "nonexistent@example.com", "password": "password123"},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name:           "wrong password",
			loginErr:       apperrors.New(apperrors.PasswordMismatch),
			requestBody:    map[string]interface{}{"email": "test@example.com", "password": "wrongpassword"},
			expectedStatus: fiber.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp(NewAuthHandler(&mockAuthUseCase{loginErr: tt.loginErr}, time.Hour))
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestAuthHandler_UsesConfiguredAccessExpiry(t *testing.T) {
	app := setupTestApp(NewAuthHandler(&mockAuthUseCase{}, 2*time.Hour))
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	var result struct {
		Data struct {
			ExpiresIn int64 `json:"expires_in"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Data.ExpiresIn != int64(2*time.Hour/time.Second) {
		t.Fatalf("expires_in = %d, want %d", result.Data.ExpiresIn, int64(2*time.Hour/time.Second))
	}
}

type mockAuthUseCase struct {
	loginErr error
}

func (m *mockAuthUseCase) Register(email, password, name string) (*User, string, string, error) {
	now := time.Now()
	return &User{ID: "generated-id", Email: email, Name: name, Role: "user", CreatedAt: now, UpdatedAt: now}, "access", "refresh", nil
}

func (m *mockAuthUseCase) Login(email, password string) (string, string, error) {
	if m.loginErr != nil {
		return "", "", m.loginErr
	}
	return "access", "refresh", nil
}

func (m *mockAuthUseCase) RefreshToken(refreshToken string) (string, string, error) {
	return "access", "refresh", nil
}

func (m *mockAuthUseCase) Logout(sessionID string) error { return nil }

func (m *mockAuthUseCase) GetProfile(userID string) (*User, error) { return nil, nil }

func (m *mockAuthUseCase) UpdateProfile(userID, name string) (*User, error) { return nil, nil }
