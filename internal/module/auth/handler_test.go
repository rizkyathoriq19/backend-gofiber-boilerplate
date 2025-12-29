package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"boilerplate-be/internal/middleware"
	apperrors "boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/security"

	"github.com/gofiber/fiber/v2"
)

// setupTestApp creates a test Fiber app with auth routes
func setupTestApp(authHandler *AuthHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Post("/register", authHandler.Register)
	app.Post("/login", authHandler.Login)

	return app
}

// TestAuthHandler_Register tests the registration endpoint
func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "valid registration",
			requestBody: map[string]interface{}{
				"email":    "newuser@example.com",
				"password": "password123",
				"name":     "New User",
			},
			expectedStatus: fiber.StatusCreated,
			expectSuccess:  true,
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
				"name":     "New User",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
				"name":  "Test User",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
				"name":     "Test User",
			},
			expectedStatus: fiber.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name:           "empty body",
			requestBody:    map[string]interface{}{},
			expectedStatus: fiber.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock dependencies
			mockRepo := NewMockAuthRepository()
			jwtManager := security.NewJWTManager("test-secret", 24*time.Hour)

			// Create a simple mock use case for testing
			// In a real scenario, you'd use the actual use case with mocked dependencies
			mockUseCase := &mockAuthUseCase{
				repo:       mockRepo,
				jwtManager: jwtManager,
			}

			handler := &AuthHandler{authUseCase: mockUseCase}
			app := setupTestApp(handler)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestAuthHandler_Login tests the login endpoint
func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockAuthRepository)
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid login",
			setupMock: func(m *MockAuthRepository) {
				hashedPassword, _ := security.HashPassword("password123")
				m.users["user-id"] = &User{
					ID:       "user-id",
					Email:    "test@example.com",
					Name:     "Test User",
					Password: hashedPassword,
					Role:     "user",
				}
			},
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:      "user not found",
			setupMock: func(m *MockAuthRepository) {},
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: "wrong password",
			setupMock: func(m *MockAuthRepository) {
				hashedPassword, _ := security.HashPassword("correctpassword")
				m.users["user-id"] = &User{
					ID:       "user-id",
					Email:    "test@example.com",
					Password: hashedPassword,
				}
			},
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: fiber.StatusUnprocessableEntity, // PasswordMismatch returns 422
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockAuthRepository()
			tt.setupMock(mockRepo)

			jwtManager := security.NewJWTManager("test-secret", 24*time.Hour)
			mockUseCase := &mockAuthUseCase{
				repo:       mockRepo,
				jwtManager: jwtManager,
			}

			handler := &AuthHandler{authUseCase: mockUseCase}
			app := setupTestApp(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to execute request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// mockAuthUseCase implements AuthUseCase for testing
type mockAuthUseCase struct {
	repo       *MockAuthRepository
	jwtManager *security.JWTManager
}

func (m *mockAuthUseCase) Register(email, password, name string) (*User, string, string, error) {
	// Check if user exists
	if _, err := m.repo.GetUserByEmail(email); err == nil {
		return nil, "", "", err
	}

	hashedPassword, err := security.HashPassword(password)
	if err != nil {
		return nil, "", "", err
	}

	user := &User{
		ID:        "generated-id",
		Email:     email,
		Password:  hashedPassword,
		Name:      name,
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.repo.CreateUser(user); err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := m.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (m *mockAuthUseCase) Login(email, password string) (string, string, error) {
	user, err := m.repo.GetUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	if err := security.CheckPassword(user.Password, password); err != nil {
		// Return proper AppError for password mismatch
		return "", "", apperrors.New(apperrors.PasswordMismatch)
	}

	return m.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role)
}

func (m *mockAuthUseCase) RefreshToken(refreshToken string) (string, string, error) {
	return "", "", nil
}

func (m *mockAuthUseCase) Logout(userID, tokenID string) error {
	return nil
}

func (m *mockAuthUseCase) GetProfile(userID string) (*User, error) {
	return m.repo.GetUserByID(userID)
}

func (m *mockAuthUseCase) UpdateProfile(userID, name string) (*User, error) {
	user, err := m.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	user.Name = name
	return user, m.repo.UpdateUser(user)
}
