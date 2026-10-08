package application

import (
	"net/http/httptest"
	"testing"
	"time"

	"boilerplate-be/internal/config"
	"boilerplate-be/internal/module/auth"
	"boilerplate-be/internal/module/rbac"
	"github.com/gofiber/fiber/v2"
)

func TestAssemblyPublicAndProtectedRoutes(t *testing.T) {
	cfg := config.New()
	cfg.RateLimit.Max = 1000
	storage := &testStorage{data: make(map[string][]byte)}
	app := New(cfg, Dependencies{Auth: auth.NewAuthHandler(&mockAuthUseCase{}, time.Hour), RBAC: rbac.NewRBACHandler(nil), RateLimitStorage: storage}, Options{})
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/ping", 200},
		{"GET", "/api/v1/auth/profile", 401},
		{"GET", "/api/v1/super-admin/roles", 401},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != tc.status {
			t.Fatalf("%s %s: got %d, want %d", tc.method, tc.path, resp.StatusCode, tc.status)
		}
	}
}

type testStorage struct{ data map[string][]byte }

func (s *testStorage) Get(key string) ([]byte, error) { return s.data[key], nil }
func (s *testStorage) Set(key string, value []byte, _ time.Duration) error {
	s.data[key] = append([]byte(nil), value...)
	return nil
}
func (s *testStorage) Delete(key string) error { delete(s.data, key); return nil }
func (s *testStorage) Reset() error            { clear(s.data); return nil }
func (s *testStorage) Close() error            { return nil }

var _ fiber.Storage = (*testStorage)(nil)
