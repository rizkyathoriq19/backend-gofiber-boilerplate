package rbac

import (
	"time"
)

// Role represents a user role in the system
type Role struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	ParentRoleID *string   `json:"parent_role_id,omitempty"`
	Level        int       `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
}

// Permission represents a permission that can be assigned to roles
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       string    `json:"role_id"`
	PermissionID string    `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// RoleWithPermissions includes role with its assigned permissions
type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}

// RoleWithChildren includes role with its child roles (hierarchical view)
type RoleWithChildren struct {
	Role
	Children []RoleWithChildren `json:"children,omitempty"`
}

// InheritedPermission represents a permission with inheritance info
type InheritedPermission struct {
	Permission
	IsInherited bool `json:"is_inherited"`
}

// UserWithRoles includes user ID with their assigned roles
type UserWithRoles struct {
	UserID string `json:"user_id"`
	Roles  []Role `json:"roles"`
}

// UserPermissions contains all permissions for a user (aggregated from all roles)
type UserPermissions struct {
	UserID      string       `json:"user_id"`
	Roles       []string     `json:"roles"`
	Permissions []Permission `json:"permissions"`
}
