package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"boilerplate-be/internal/shared/enum"
	apperrors "boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/security"

	"github.com/gofiber/fiber/v2"
)

type authMiddlewareSessionStore struct {
	mu   sync.Mutex
	data map[string]string
	err  error
}

func (s *authMiddlewareSessionStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return "", s.err
	}
	return s.data[key], nil
}

func (s *authMiddlewareSessionStore) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.data[key] = value
	return nil
}

func (s *authMiddlewareSessionStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	delete(s.data, key)
	return nil
}

func (s *authMiddlewareSessionStore) CompareAndSwap(_ context.Context, key, expected, replacement string, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return false, s.err
	}
	if s.data[key] != expected {
		return false, nil
	}
	s.data[key] = replacement
	return true, nil
}

func TestAuthMiddleware_Returns503WhenSessionStoreUnavailable(t *testing.T) {
	store := &authMiddlewareSessionStore{data: make(map[string]string)}
	manager := security.NewLoginSessionManager(
		security.NewJWTManager("secret", time.Hour, time.Hour), store,
	)
	pair, err := manager.Issue("user-1", "one@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}
	store.err = errors.New("redis unavailable")

	app := fiber.New()
	app.Use(AuthMiddleware(manager))
	app.Get("/protected", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
	}

	var body struct {
		ErrorCode int `json:"error_code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ErrorCode != apperrors.AuthServiceUnavailable.Value() {
		t.Fatalf("error_code = %d, want %d", body.ErrorCode, apperrors.AuthServiceUnavailable.Value())
	}
}
