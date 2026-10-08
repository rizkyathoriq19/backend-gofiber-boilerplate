package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"boilerplate-be/internal/config"
	"boilerplate-be/internal/delivery/websocket"
	"boilerplate-be/internal/module/auth"
	"boilerplate-be/internal/module/rbac"
	"boilerplate-be/internal/shared/enum"
	apperrors "boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/security"
	"github.com/gofiber/fiber/v2"
)

func request(t *testing.T, app *fiber.App, method, path, body, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, status int) {
	t.Helper()
	if resp.StatusCode != status {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d; %s", resp.StatusCode, status, body)
	}
}

func assertError(t *testing.T, resp *http.Response, status int, code enum.ErrorCode) {
	t.Helper()
	assertStatus(t, resp, status)
	var body response.BaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status || body.Code != status || body.ErrorCode != code.Value() || body.Message != code.MessageEN() || body.Timestamp.IsZero() {
		t.Fatalf("unexpected error envelope: %+v", body)
	}
}

func TestAssemblyLoginSessionAndAuthorization(t *testing.T) {
	cfg := config.New()
	cfg.RateLimit.Max = 1000
	store := &sessionStore{data: map[string]string{}}
	sessions := security.NewLoginSessionManager(security.NewJWTManager("test-secret", time.Hour, time.Hour), store)
	useCase := auth.NewAuthUseCase(&accountRepository{}, sessions)
	authorization := &roleUseCase{}
	app := New(cfg, Dependencies{Auth: auth.NewAuthHandler(useCase, time.Hour), RBAC: rbac.NewRBACHandler(authorization), Authorization: authorization, Sessions: sessions, RateLimitStorage: &testStorage{data: map[string][]byte{}}}, Options{})
	// Public registration issues credentials without authentication middleware.
	resp := request(t, app, "POST", "/api/v1/auth/register", `{"email":"user@example.com","password":"password123","name":"User"}`, "")
	assertStatus(t, resp, 201)
	var registered struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&registered); err != nil {
		t.Fatal(err)
	}
	token := registered.Data.AccessToken
	if token == "" {
		t.Fatal("registration did not issue access token")
	}
	assertStatus(t, request(t, app, "GET", "/api/v1/auth/profile", "", token), 200)
	assertStatus(t, request(t, app, "PUT", "/api/v1/auth/profile", `{"name":"Updated"}`, token), 200)
	assertStatus(t, request(t, app, "GET", "/api/v1/auth/my-roles", "", token), 200)
	assertStatus(t, request(t, app, "GET", "/api/v1/auth/my-permissions", "", token), 200)
	for _, path := range []string{"/api/v1/auth/profile", "/api/v1/super-admin/roles"} {
		assertError(t, request(t, app, "GET", path, "", ""), 401, apperrors.Unauthorized)
		assertError(t, request(t, app, "GET", path, "", "invalid"), 401, apperrors.InvalidToken)
	}
	// Session validation must precede role checks; token role alone grants nothing.
	assertError(t, request(t, app, "GET", "/api/v1/super-admin/roles", "", token), 403, apperrors.Forbidden)
	authorization.allowed = true
	assertStatus(t, request(t, app, "GET", "/api/v1/super-admin/roles", "", token), 200)
	if authorization.userID == "" {
		t.Fatal("authorization did not receive authenticated user")
	}
	authorization.err = errors.New("role store unavailable")
	assertError(t, request(t, app, "GET", "/api/v1/super-admin/roles", "", token), 500, apperrors.InternalServerError)
	authorization.err = nil
	store.err = errors.New("session store unavailable")
	assertError(t, request(t, app, "GET", "/api/v1/auth/profile", "", token), 503, apperrors.AuthServiceUnavailable)
	assertError(t, request(t, app, "GET", "/api/v1/super-admin/roles", "", token), 503, apperrors.AuthServiceUnavailable)
	store.err = nil
	// Renew through the public production route; consumed credentials cannot renew again.
	refreshBody := `{"refresh_token":"` + registered.Data.RefreshToken + `"}`
	assertStatus(t, request(t, app, "POST", "/api/v1/auth/refresh", refreshBody, ""), 200)
	assertStatus(t, request(t, app, "POST", "/api/v1/auth/refresh", refreshBody, ""), 401)
	assertStatus(t, request(t, app, "POST", "/api/v1/auth/logout", "", token), 200)
	assertError(t, request(t, app, "GET", "/api/v1/auth/profile", "", token), 401, apperrors.InvalidToken)
}

func TestAssemblyOptionalRoutes(t *testing.T) {
	t.Chdir("../..") // Match the production working directory for static documentation.
	for _, options := range []Options{{}, {Docs: true}, {Presentation: true}, {Docs: true, Presentation: true}} {
		for _, includeWS := range []bool{false, true} {
			cfg := config.New()
			cfg.RateLimit.Max = 1000
			deps := Dependencies{Auth: auth.NewAuthHandler(&mockAuthUseCase{}, time.Hour), RBAC: rbac.NewRBACHandler(nil), RateLimitStorage: &testStorage{data: map[string][]byte{}}}
			if includeWS {
				deps.WebSocket = websocket.NewHub()
			} // No worker is needed to assemble or reject a non-upgrade request.
			app := New(cfg, deps, options)
			for _, path := range []string{"/", "/api/v1/health"} {
				status := 404
				if options.Presentation {
					status = 200
				}
				assertStatus(t, request(t, app, "GET", path, "", ""), status)
			}
			for _, path := range []string{"/swagger/index.html", "/swagger/doc.json", "/docs/swagger.json"} {
				status := 404
				if options.Docs {
					status = 200
				}
				assertStatus(t, request(t, app, "GET", path, "", ""), status)
			}
			status := 404
			if includeWS {
				status = 500
			} // Preserve current centralized handling of Fiber's upgrade error.
			assertStatus(t, request(t, app, "GET", "/ws/", "", ""), status)
			assertError(t, request(t, app, "GET", "/missing", "", ""), 404, apperrors.ResourceNotFound)
			req := httptest.NewRequest("GET", "/missing", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			want := "application/json"
			if options.Presentation {
				want = "text/html"
			}
			if !bytes.Contains([]byte(resp.Header.Get("Content-Type")), []byte(want)) {
				t.Fatalf("fallback content type: %s", resp.Header.Get("Content-Type"))
			}
		}
	}
}

func TestAssemblyMiddlewareOrderAndErrors(t *testing.T) {
	cfg := config.New()
	cfg.RateLimit.Max = 1
	cfg.CORS.AllowedOrigins = []string{"https://client.example"}
	app := New(cfg, Dependencies{Auth: auth.NewAuthHandler(&mockAuthUseCase{}, time.Hour), RBAC: rbac.NewRBACHandler(nil), RateLimitStorage: &testStorage{data: map[string][]byte{}}}, Options{})
	preflight := httptest.NewRequest("OPTIONS", "/api/v1/auth/profile", nil)
	preflight.Header.Set("Origin", "https://client.example")
	preflight.Header.Set("Access-Control-Request-Method", "GET")
	preflight.Header.Set("X-Request-ID", "preflight-id")
	resp, err := app.Test(preflight)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp, 204)
	if resp.Header.Get("X-Request-ID") != "preflight-id" || resp.Header.Get("Access-Control-Allow-Origin") != "https://client.example" {
		t.Fatal("request ID must precede CORS")
	}
	// CORS preflight must not consume limiter budget; ETag wraps the public response.
	ping := request(t, app, "GET", "/ping", "", "")
	assertStatus(t, ping, 200)
	if ping.Header.Get("Etag") == "" {
		t.Fatal("missing ETag")
	}
	limited := request(t, app, "GET", "/api/v1/auth/profile", "", "")
	assertError(t, limited, 429, apperrors.RateLimitExceeded)
	if limited.Header.Get("X-Request-ID") == "" || limited.Header.Get("X-Content-Type-Options") == "" {
		t.Fatal("request ID and helmet must precede limiter")
	}
	cfg.RateLimit.Max = 1000
	sessions := security.NewLoginSessionManager(security.NewJWTManager("secret", time.Hour, time.Hour), &sessionStore{data: map[string]string{}})
	pair, err := sessions.Issue("user", "user@example.com", enum.UserRoleUser)
	if err != nil {
		t.Fatal(err)
	}
	app = New(cfg, Dependencies{Auth: auth.NewAuthHandler(&mockAuthUseCase{loginErr: errors.New("unexpected")}, time.Hour), RBAC: rbac.NewRBACHandler(nil), Sessions: sessions, RateLimitStorage: &testStorage{data: map[string][]byte{}}}, Options{})
	assertError(t, request(t, app, "POST", "/api/v1/auth/login", `{"email":"user@example.com","password":"password123"}`, ""), 500, apperrors.InternalServerError)
	// The handler mock deliberately returns a nil profile; recover must catch its panic.
	assertError(t, request(t, app, "GET", "/api/v1/auth/profile", "", pair.AccessToken), 500, apperrors.InternalServerError)
	assertError(t, request(t, app, "POST", "/api/v1/auth/login", `{`, ""), 400, apperrors.InvalidRequestBody)
}

type accountRepository struct{ user *auth.User }

func (r *accountRepository) CreateUser(user *auth.User) error {
	user.ID = "user-1"
	user.Role = enum.UserRoleUser
	r.user = user
	return nil
}
func (r *accountRepository) GetUserByEmail(string) (*auth.User, error) {
	if r.user == nil {
		return nil, apperrors.New(apperrors.AccountNotFound)
	}
	return r.user, nil
}
func (r *accountRepository) GetUserByID(string) (*auth.User, error) { return r.user, nil }
func (r *accountRepository) UpdateUser(user *auth.User) error       { r.user = user; return nil }

type roleUseCase struct {
	rbac.RBACUseCase
	allowed bool
	userID  string
	err     error
}

func (r *roleUseCase) CheckUserRole(userID string, _ ...string) (bool, error) {
	r.userID = userID
	return r.allowed, r.err
}
func (r *roleUseCase) GetRoles() ([]rbac.Role, error)           { return []rbac.Role{}, nil }
func (r *roleUseCase) GetUserRoles(string) ([]rbac.Role, error) { return []rbac.Role{}, nil }
func (r *roleUseCase) GetUserPermissions(string) ([]rbac.Permission, error) {
	return []rbac.Permission{}, nil
}

type sessionStore struct {
	mu   sync.Mutex
	data map[string]string
	err  error
}

func (s *sessionStore) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[key], s.err
}
func (s *sessionStore) Set(_ context.Context, key, value string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	s.data[key] = value
	return nil
}
func (s *sessionStore) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	delete(s.data, key)
	return nil
}
func (s *sessionStore) CompareAndSwap(_ context.Context, key, expected, replacement string, _ time.Duration) (bool, error) {
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
