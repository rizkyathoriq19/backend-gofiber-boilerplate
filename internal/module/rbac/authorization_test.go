package rbac_test

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"boilerplate-be/internal/middleware"
	"boilerplate-be/internal/module/rbac"
	apiresponse "boilerplate-be/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Each test owns a schema; never mutate application tables.
func authorizationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("RBAC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration unavailable: set RBAC_TEST_DATABASE_URL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	schema := "rbac_test_" + uuid.New().String()
	execSQL(t, db, `CREATE SCHEMA "`+schema+`"`)
	t.Cleanup(func() { execSQL(t, db, `DROP SCHEMA "`+schema+`" CASCADE`) })
	connectionURL, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parameters := connectionURL.Query()
	parameters.Set("search_path", `"`+schema+`"`)
	connectionURL.RawQuery = parameters.Encode()
	testDB, err := sql.Open("postgres", connectionURL.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { testDB.Close() })
	// Only the RBAC migration is needed; users are owned by authentication.
	execSQL(t, testDB, `CREATE TABLE users (id UUID PRIMARY KEY)`)
	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "20251216020240_create_rbac_tables.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	execSQL(t, testDB, string(migration))
	return testDB
}

func execSQL(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}

func TestAuthorizationReflectsCommittedMutations(t *testing.T) {
	db := authorizationDB(t)
	repo := rbac.NewRBACRepository(db)
	service := rbac.NewRBACUseCase(repo)
	userID := uuid.New().String()
	execSQL(t, db, `INSERT INTO users (id) VALUES ($1)`, userID)
	role, err := service.CreateRole("editor", "")
	if err != nil {
		t.Fatal(err)
	}
	permission, err := repo.GetPermissionByName("users:read")
	if err != nil {
		t.Fatal(err)
	}
	check := func(roleName string, wantRole, wantPermission bool) {
		t.Helper()
		got, err := service.CheckUserRole(userID, "missing", roleName)
		if err != nil || got != wantRole {
			t.Fatalf("role: got %v, %v; want %v", got, err, wantRole)
		}
		got, err = service.CheckUserPermission(userID, "missing", "users:read")
		if err != nil || got != wantPermission {
			t.Fatalf("permission: got %v, %v; want %v", got, err, wantPermission)
		}
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	check("editor", false, false)
	must(service.AssignRoleToUser(userID, role.ID))
	check("editor", true, false)
	must(service.AssignPermissionToRole(role.ID, permission.ID))
	check("editor", true, true) // Repeated grants used to populate authorization caches.
	check("editor", true, true)
	must(service.RemovePermissionFromRole(role.ID, permission.ID))
	check("editor", true, false)
	must(service.AssignPermissionToRole(role.ID, permission.ID))
	check("editor", true, true)
	must(service.RemoveRoleFromUser(userID, role.ID))
	check("editor", false, false)
	must(service.AssignRoleToUser(userID, role.ID))
	check("editor", true, true)
	_, err = service.UpdateRole(role.ID, "renamed", "")
	must(err)
	check("editor", false, true)
	check("renamed", true, true)
	must(service.DeleteRole(role.ID))
	check("renamed", false, false)
	for _, name := range []string{"user", "super_admin"} {
		system, err := repo.GetRoleByName(name)
		must(err)
		if err := service.DeleteRole(system.ID); err == nil {
			t.Fatalf("system role %s was deleted", name)
		}
	}
}

func TestAuthorizationMiddlewareAnyAllAndStorageFailure(t *testing.T) {
	db := authorizationDB(t)
	repo := rbac.NewRBACRepository(db)
	service := rbac.NewRBACUseCase(repo)
	userID := uuid.New().String()
	execSQL(t, db, `INSERT INTO users (id) VALUES ($1)`, userID)
	role, err := repo.GetRoleByName("user")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AssignRoleToUser(userID, role.ID); err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user_id", userID); return c.Next() })
	routes := map[string]fiber.Handler{
		"/any-role":               middleware.RequireAnyRole(service, "missing", "user"),
		"/all-role":               middleware.RequireAllRoles(service, "user", "admin"),
		"/all-present-role":       middleware.RequireAllRoles(service, "user"),
		"/any-permission":         middleware.RequirePermission(service, "missing", "profile:read"),
		"/all-permission":         middleware.RequireAllPermissions(service, "profile:read", "users:read"),
		"/all-present-permission": middleware.RequireAllPermissions(service, "profile:read", "profile:write"),
	}
	for path, handler := range routes {
		app.Get(path, handler, func(c *fiber.Ctx) error { return c.SendStatus(204) })
	}
	check := func(path string, status int) {
		t.Helper()
		response, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != status {
			t.Fatalf("%s: got %d, want %d", path, response.StatusCode, status)
		}
		if status >= 400 {
			var envelope apiresponse.BaseResponse
			if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Status || envelope.Code != status || envelope.ErrorCode == 0 || envelope.Message == "" || envelope.Timestamp.IsZero() {
				t.Fatalf("%s: invalid error envelope: %+v", path, envelope)
			}
		}
	}
	for _, path := range []string{"/any-role", "/all-present-role", "/any-permission", "/all-present-permission"} {
		check(path, 204)
	}
	check("/all-role", 403)
	check("/all-permission", 403)
	admin, err := repo.GetRoleByName("admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := service.AssignRoleToUser(userID, admin.ID); err != nil {
		t.Fatal(err)
	}
	check("/all-role", 204)
	check("/all-permission", 204)
	if err := service.RemoveRoleFromUser(userID, admin.ID); err != nil {
		t.Fatal(err)
	}
	check("/all-role", 403)
	check("/all-permission", 403)
	if err := service.RemoveRoleFromUser(userID, role.ID); err != nil {
		t.Fatal(err)
	}
	for path := range routes {
		check(path, 403)
	}
	if err := service.AssignRoleToUser(userID, role.ID); err != nil {
		t.Fatal(err)
	}
	check("/any-role", 204)
	check("/any-permission", 204)
	// Previously warm grants cannot hide a PostgreSQL lookup failure.
	execSQL(t, db, `ALTER TABLE user_roles RENAME TO unavailable_user_roles`)
	for path := range routes {
		check(path, 500)
	}
	if granted, err := service.CheckUserRole(userID, "user"); granted || err == nil {
		t.Fatalf("storage failure granted role: %v, %v", granted, err)
	}
	if granted, err := service.CheckUserPermission(userID, "profile:read"); granted || err == nil {
		t.Fatalf("storage failure granted permission: %v, %v", granted, err)
	}
}
