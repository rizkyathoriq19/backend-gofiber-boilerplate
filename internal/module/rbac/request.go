package rbac

// CreateRoleRequest is the request body for creating a new role
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"max=255"`
}

// UpdateRoleRequest is the request body for updating a role
type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=50"`
	Description string `json:"description" validate:"max=255"`
}

// AssignRoleRequest is the request body for assigning a role to a user
type AssignRoleRequest struct {
	RoleID string `json:"role_id" validate:"required,uuid"`
}

// AssignPermissionRequest is the request body for assigning a permission to a role
type AssignPermissionRequest struct {
	PermissionID string `json:"permission_id" validate:"required,uuid"`
}

// BatchAssignPermissionsRequest is the request body for batch assigning permissions to a role
type BatchAssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1,dive,uuid"`
}

// BatchRemovePermissionsRequest is the request body for batch removing permissions from a role
type BatchRemovePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1,dive,uuid"`
}

// BatchGetRolePermissionsRequest is the request body for batch getting permissions by roles
type BatchGetRolePermissionsRequest struct {
	RoleIDs []string `json:"role_ids" validate:"required,min=1,dive,uuid"`
}
