package rbac

import "time"

// RoleResponse is the response for a single role
type RoleResponse struct {
	ID          string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PermissionResponse is the response for a single permission
type PermissionResponse struct {
	ID          string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleWithPermissionsResponse includes role with its permissions
type RoleWithPermissionsResponse struct {
	RoleResponse
	Permissions []PermissionResponse `json:"permissions"`
}

// UserRolesResponse contains user's roles
type UserRolesResponse struct {
	UserID string         `json:"user_uuid"`
	Roles  []RoleResponse `json:"roles"`
}

// BatchRolePermissionsResponse contains permissions grouped by role
type BatchRolePermissionsResponse struct {
	RolePermissions map[string][]PermissionResponse `json:"role_permissions"`
}

// ToRoleResponse converts Role entity to RoleResponse
func ToRoleResponse(role *Role) RoleResponse {
	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}

// ToRoleResponses converts slice of Role entities to slice of RoleResponse
func ToRoleResponses(roles []Role) []RoleResponse {
	result := make([]RoleResponse, len(roles))
	for i, role := range roles {
		result[i] = ToRoleResponse(&role)
	}
	return result
}

// ToPermissionResponse converts Permission entity to PermissionResponse
func ToPermissionResponse(permission *Permission) PermissionResponse {
	return PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		Resource:    permission.Resource,
		Action:      permission.Action,
		CreatedAt:   permission.CreatedAt,
	}
}

// ToPermissionResponses converts slice of Permission entities to slice of PermissionResponse
func ToPermissionResponses(permissions []Permission) []PermissionResponse {
	result := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		result[i] = ToPermissionResponse(&permission)
	}
	return result
}
