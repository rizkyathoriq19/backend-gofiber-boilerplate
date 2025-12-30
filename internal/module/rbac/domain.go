package rbac

// RBACRepository defines the data access layer for RBAC operations
type RBACRepository interface {
	// Role operations
	GetRoles() ([]Role, error)
	GetRoleByID(id string) (*Role, error)
	GetRoleByName(name string) (*Role, error)
	CreateRole(role *Role) error
	UpdateRole(role *Role) error
	DeleteRole(id string) error
	GetRoleHierarchy() ([]RoleWithChildren, error)
	GetRoleAncestors(roleID string) ([]Role, error)
	GetRoleDescendants(roleID string) ([]Role, error)

	// Permission operations
	GetPermissions() ([]Permission, error)
	GetPermissionByID(id string) (*Permission, error)
	GetPermissionByName(name string) (*Permission, error)
	CreatePermission(permission *Permission) error

	// User-Role operations
	GetUserRoles(userID string) ([]Role, error)
	AssignRoleToUser(userID, roleID string) error
	RemoveRoleFromUser(userID, roleID string) error
	HasRole(userID, roleName string) (bool, error)

	// Role-Permission operations
	GetRolePermissions(roleID string) ([]Permission, error)
	GetRolePermissionsWithInheritance(roleID string) ([]InheritedPermission, error)
	AssignPermissionToRole(roleID, permissionID string) error
	RemovePermissionFromRole(roleID, permissionID string) error

	// Hierarchical role operations
	SetParentRole(roleID, parentRoleID string) error

	// User permission check (aggregated from all user's roles)
	GetUserPermissions(userID string) ([]Permission, error)
	HasPermission(userID, permissionName string) (bool, error)
}

// RBACUseCase defines the business logic for RBAC operations
type RBACUseCase interface {
	// Role operations
	GetRoles() ([]Role, error)
	GetRoleByID(id string) (*Role, error)
	CreateRole(name, description string) (*Role, error)
	UpdateRole(id, name, description string) (*Role, error)
	DeleteRole(id string) error

	// Permission operations
	GetPermissions() ([]Permission, error)

	// User-Role operations
	GetUserRoles(userID string) ([]Role, error)
	AssignRoleToUser(userID, roleID string) error
	RemoveRoleFromUser(userID, roleID string) error

	// Role-Permission operations
	GetRolePermissions(roleID string) ([]Permission, error)
	GetRolePermissionsWithInheritance(roleID string) ([]InheritedPermission, error)
	AssignPermissionToRole(roleID, permissionID string) error
	RemovePermissionFromRole(roleID, permissionID string) error

	// Role hierarchy operations
	GetRoleHierarchy() ([]RoleWithChildren, error)
	SetParentRole(roleID, parentRoleID string) error

	// Permission checking
	CheckUserRole(userID string, roles ...string) (bool, error)
	CheckUserPermission(userID string, permissions ...string) (bool, error)
	GetUserPermissions(userID string) ([]Permission, error)
}
