package rbac_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"boilerplate-be/internal/module/rbac"
)

// Inject a failure after a valid grant: a query error alone cannot catch
// returning partial results without checking rows.Err().
type interruptedDriver struct{}
type interruptedConnection struct{}
type interruptedRows struct {
	permission bool
	delivered  bool
}

func (interruptedDriver) Open(string) (driver.Conn, error) { return interruptedConnection{}, nil }
func (interruptedConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unsupported")
}
func (interruptedConnection) Close() error              { return nil }
func (interruptedConnection) Begin() (driver.Tx, error) { return nil, errors.New("unsupported") }
func (interruptedConnection) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	return &interruptedRows{permission: strings.Contains(query, "FROM permissions")}, nil
}
func (r *interruptedRows) Columns() []string {
	if r.permission {
		return []string{"id", "name", "description", "resource", "action", "created_at"}
	}
	return []string{"id", "name", "description", "created_at"}
}
func (*interruptedRows) Close() error { return nil }
func (r *interruptedRows) Next(values []driver.Value) error {
	if r.delivered {
		return errors.New("interrupted SQL iteration")
	}
	r.delivered = true
	row := []driver.Value{"id", "user", "description", time.Now()}
	if r.permission {
		row = []driver.Value{"id", "profile:read", "description", "profile", "read", time.Now()}
	}
	copy(values, row)
	return nil
}

func init() { sql.Register("rbac_interrupted_rows", interruptedDriver{}) }

func TestAuthorizationRejectsPartialSQLResults(t *testing.T) {
	db, err := sql.Open("rbac_interrupted_rows", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := rbac.NewRBACRepository(db)
	service := rbac.NewRBACUseCase(repo)
	for name, check := range map[string]func() (bool, error){
		"role":                  func() (bool, error) { return service.CheckUserRole("user-id", "user") },
		"permission":            func() (bool, error) { return service.CheckUserPermission("user-id", "profile:read") },
		"repository role":       func() (bool, error) { return repo.HasRole("user-id", "user") },
		"repository permission": func() (bool, error) { return repo.HasPermission("user-id", "profile:read") },
	} {
		t.Run(name, func(t *testing.T) {
			if granted, err := check(); granted || err == nil {
				t.Fatalf("partial results: grant=%v, error=%v", granted, err)
			}
		})
	}
	if roles, err := service.GetRoles(); roles != nil || err == nil {
		t.Fatalf("partial roles: %v, %v", roles, err)
	}
	if permissions, err := service.GetPermissions(); permissions != nil || err == nil {
		t.Fatalf("partial permissions: %v, %v", permissions, err)
	}
	if permissions, err := repo.GetRolePermissions("role-id"); permissions != nil || err == nil {
		t.Fatalf("partial role permissions: %v, %v", permissions, err)
	}
}
