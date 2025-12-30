package rbac

import (
	"boilerplate-be/internal/pkg/errors"
)

type rbacUseCase struct {
	rbacRepo RBACRepository
}

// NewRBACUseCase creates a new RBAC use case
func NewRBACUseCase(rbacRepo RBACRepository) RBACUseCase {
	return &rbacUseCase{
		rbacRepo: rbacRepo,
	}
}

// ==================== Role Operations ====================

func (u *rbacUseCase) GetRoles() ([]Role, error) {
	return u.rbacRepo.GetRoles()
}

func (u *rbacUseCase) GetRoleByID(id string) (*Role, error) {
	return u.rbacRepo.GetRoleByID(id)
}

func (u *rbacUseCase) CreateRole(name, description string) (*Role, error) {
	// Check if role already exists
	if _, err := u.rbacRepo.GetRoleByName(name); err == nil {
		return nil, errors.New(errors.Conflict)
	}

	role := &Role{
		Name:        name,
		Description: description,
	}

	if err := u.rbacRepo.CreateRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (u *rbacUseCase) UpdateRole(id, name, description string) (*Role, error) {
	role, err := u.rbacRepo.GetRoleByID(id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}

	if err := u.rbacRepo.UpdateRole(role); err != nil {
		return nil, err
	}

	return role, nil
}

func (u *rbacUseCase) DeleteRole(id string) error {
	// Check if role exists
	role, err := u.rbacRepo.GetRoleByID(id)
	if err != nil {
		return err
	}

	// Prevent deletion of system roles
	if role.Name == "super_admin" || role.Name == "user" {
		return errors.New(errors.Forbidden)
	}

	return u.rbacRepo.DeleteRole(id)
}

// ==================== Permission Operations ====================

func (u *rbacUseCase) GetPermissions() ([]Permission, error) {
	return u.rbacRepo.GetPermissions()
}

// ==================== User-Role Operations ====================

func (u *rbacUseCase) GetUserRoles(userID string) ([]Role, error) {
	return u.rbacRepo.GetUserRoles(userID)
}

func (u *rbacUseCase) AssignRoleToUser(userID, roleID string) error {
	// Verify role exists
	if _, err := u.rbacRepo.GetRoleByID(roleID); err != nil {
		return err
	}

	return u.rbacRepo.AssignRoleToUser(userID, roleID)
}

func (u *rbacUseCase) RemoveRoleFromUser(userID, roleID string) error {
	return u.rbacRepo.RemoveRoleFromUser(userID, roleID)
}

// ==================== Role-Permission Operations ====================

func (u *rbacUseCase) GetRolePermissions(roleID string) ([]Permission, error) {
	// Verify role exists
	if _, err := u.rbacRepo.GetRoleByID(roleID); err != nil {
		return nil, err
	}

	return u.rbacRepo.GetRolePermissions(roleID)
}

func (u *rbacUseCase) AssignPermissionToRole(roleID, permissionID string) error {
	// Verify role exists
	if _, err := u.rbacRepo.GetRoleByID(roleID); err != nil {
		return err
	}

	// Verify permission exists
	if _, err := u.rbacRepo.GetPermissionByID(permissionID); err != nil {
		return err
	}

	return u.rbacRepo.AssignPermissionToRole(roleID, permissionID)
}

func (u *rbacUseCase) RemovePermissionFromRole(roleID, permissionID string) error {
	return u.rbacRepo.RemovePermissionFromRole(roleID, permissionID)
}

// ==================== Permission Checking ====================

func (u *rbacUseCase) CheckUserRole(userID string, roles ...string) (bool, error) {
	userRoles, err := u.rbacRepo.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role.Name] = true
	}

	for _, required := range roles {
		if roleMap[required] {
			return true, nil
		}
	}

	return false, nil
}

func (u *rbacUseCase) CheckUserPermission(userID string, permissions ...string) (bool, error) {
	userPermissions, err := u.rbacRepo.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	permissionMap := make(map[string]bool)
	for _, permission := range userPermissions {
		permissionMap[permission.Name] = true
	}

	for _, required := range permissions {
		if permissionMap[required] {
			return true, nil
		}
	}

	return false, nil
}

func (u *rbacUseCase) GetUserPermissions(userID string) ([]Permission, error) {
	return u.rbacRepo.GetUserPermissions(userID)
}

