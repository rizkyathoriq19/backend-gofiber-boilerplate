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
