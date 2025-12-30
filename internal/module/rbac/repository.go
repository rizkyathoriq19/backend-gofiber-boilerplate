package rbac

import (
	"context"
	"database/sql"
	"time"

	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/utils"

	"github.com/google/uuid"
)

type rbacRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewRBACRepository creates a new RBAC repository
func NewRBACRepository(db *sql.DB, cacheHelper *utils.CacheHelper) RBACRepository {
	return &rbacRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// ==================== Role Operations ====================

func (r *rbacRepository) GetRoles() ([]Role, error) {
	query := `SELECT id, name, description, parent_role_id, COALESCE(level, 0), created_at FROM roles ORDER BY level, name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var description sql.NullString
		var parentRoleID sql.NullString
		if err := rows.Scan(&role.ID, &role.Name, &description, &parentRoleID, &role.Level, &role.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		role.Description = description.String
		if parentRoleID.Valid {
			role.ParentRoleID = &parentRoleID.String
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *rbacRepository) GetRoleByID(id string) (*Role, error) {
	query := `SELECT id, name, description, parent_role_id, COALESCE(level, 0), created_at FROM roles WHERE id = $1`
	var role Role
	var description sql.NullString
	var parentRoleID sql.NullString
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &description, &parentRoleID, &role.Level, &role.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	role.Description = description.String
	if parentRoleID.Valid {
		role.ParentRoleID = &parentRoleID.String
	}
	return &role, nil
}

func (r *rbacRepository) GetRoleByName(name string) (*Role, error) {
	query := `SELECT id, name, description, parent_role_id, COALESCE(level, 0), created_at FROM roles WHERE name = $1`
	var role Role
	var description sql.NullString
	var parentRoleID sql.NullString
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &description, &parentRoleID, &role.Level, &role.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	role.Description = description.String
	if parentRoleID.Valid {
		role.ParentRoleID = &parentRoleID.String
	}
	return &role, nil
}

func (r *rbacRepository) CreateRole(role *Role) error {
	role.ID = uuid.New().String()
	role.CreatedAt = time.Now()

	query := `INSERT INTO roles (id, name, description, parent_role_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, role.ID, role.Name, role.Description, role.ParentRoleID, role.Level, role.CreatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}
	return nil
}

func (r *rbacRepository) UpdateRole(role *Role) error {
	query := `UPDATE roles SET name = $2, description = $3, parent_role_id = $4, level = $5 WHERE id = $1`
	result, err := r.db.Exec(query, role.ID, role.Name, role.Description, role.ParentRoleID, role.Level)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}
	return nil
}

func (r *rbacRepository) DeleteRole(id string) error {
	query := `DELETE FROM roles WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}
	return nil
}

// ==================== Permission Operations ====================

func (r *rbacRepository) GetPermissions() ([]Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions ORDER BY resource, action`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var permission Permission
		var description sql.NullString
		if err := rows.Scan(&permission.ID, &permission.Name, &description, &permission.Resource, &permission.Action, &permission.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		permission.Description = description.String
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (r *rbacRepository) GetPermissionByID(id string) (*Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE id = $1`
	var permission Permission
	var description sql.NullString
	err := r.db.QueryRow(query, id).Scan(&permission.ID, &permission.Name, &description, &permission.Resource, &permission.Action, &permission.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	permission.Description = description.String
	return &permission, nil
}

func (r *rbacRepository) GetPermissionByName(name string) (*Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at FROM permissions WHERE name = $1`
	var permission Permission
	var description sql.NullString
	err := r.db.QueryRow(query, name).Scan(&permission.ID, &permission.Name, &description, &permission.Resource, &permission.Action, &permission.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	permission.Description = description.String
	return &permission, nil
}

func (r *rbacRepository) CreatePermission(permission *Permission) error {
	permission.ID = uuid.New().String()
	permission.CreatedAt = time.Now()

	query := `INSERT INTO permissions (id, name, description, resource, action, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, permission.ID, permission.Name, permission.Description, permission.Resource, permission.Action, permission.CreatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}
	return nil
}

// ==================== User-Role Operations ====================

func (r *rbacRepository) GetUserRoles(userID string) ([]Role, error) {
	cacheKey := r.cacheHelper.BuildUserCacheKey(userID, "roles")

	// Try cache first
	var cachedRoles []Role
	if err := r.cacheHelper.GetJSON(context.Background(), cacheKey, &cachedRoles); err == nil {
		return cachedRoles, nil
	}

	query := `
		SELECT r.id, r.name, r.description, r.created_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var description sql.NullString
		if err := rows.Scan(&role.ID, &role.Name, &description, &role.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		role.Description = description.String
		roles = append(roles, role)
	}

	// Cache the result
	_ = r.cacheHelper.CacheJSON(context.Background(), cacheKey, roles, 5*time.Minute)

	return roles, nil
}

func (r *rbacRepository) AssignRoleToUser(userID, roleID string) error {
	query := `INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, userID, roleID, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	// Invalidate cache
	_ = r.cacheHelper.InvalidateUserCache(context.Background(), userID)
	return nil
}

func (r *rbacRepository) RemoveRoleFromUser(userID, roleID string) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.Exec(query, userID, roleID)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	// Invalidate cache
	_ = r.cacheHelper.InvalidateUserCache(context.Background(), userID)
	return nil
}

func (r *rbacRepository) HasRole(userID, roleName string) (bool, error) {
	roles, err := r.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == roleName {
			return true, nil
		}
	}
	return false, nil
}

// ==================== Role-Permission Operations ====================

func (r *rbacRepository) GetRolePermissions(roleID string) ([]Permission, error) {
	query := `
		SELECT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var permission Permission
		var description sql.NullString
		if err := rows.Scan(&permission.ID, &permission.Name, &description, &permission.Resource, &permission.Action, &permission.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		permission.Description = description.String
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (r *rbacRepository) AssignPermissionToRole(roleID, permissionID string) error {
	query := `INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, roleID, permissionID, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}
	return nil
}

func (r *rbacRepository) RemovePermissionFromRole(roleID, permissionID string) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	_, err := r.db.Exec(query, roleID, permissionID)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}
	return nil
}

// ==================== User Permission Check ====================

func (r *rbacRepository) GetUserPermissions(userID string) ([]Permission, error) {
	cacheKey := r.cacheHelper.BuildUserCacheKey(userID, "permissions")

	// Try cache first
	var cachedPermissions []Permission
	if err := r.cacheHelper.GetJSON(context.Background(), cacheKey, &cachedPermissions); err == nil {
		return cachedPermissions, nil
	}

	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY p.resource, p.action
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var permission Permission
		var description sql.NullString
		if err := rows.Scan(&permission.ID, &permission.Name, &description, &permission.Resource, &permission.Action, &permission.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		permission.Description = description.String
		permissions = append(permissions, permission)
	}

	// Cache the result
	_ = r.cacheHelper.CacheJSON(context.Background(), cacheKey, permissions, 5*time.Minute)

	return permissions, nil
}

func (r *rbacRepository) HasPermission(userID, permissionName string) (bool, error) {
	permissions, err := r.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, permission := range permissions {
		if permission.Name == permissionName {
			return true, nil
		}
	}
	return false, nil
}

// ==================== Hierarchical Role Operations ====================

func (r *rbacRepository) GetRoleHierarchy() ([]RoleWithChildren, error) {
	roles, err := r.GetRoles()
	if err != nil {
		return nil, err
	}

	// Build a map for quick lookup
	roleMap := make(map[string]*RoleWithChildren)
	for _, role := range roles {
		roleMap[role.ID] = &RoleWithChildren{Role: role}
	}

	// Build hierarchy
	var rootRoles []RoleWithChildren
	for _, role := range roles {
		if role.ParentRoleID == nil {
			rootRoles = append(rootRoles, *roleMap[role.ID])
		} else {
			if parent, ok := roleMap[*role.ParentRoleID]; ok {
				parent.Children = append(parent.Children, *roleMap[role.ID])
			}
		}
	}

	// Rebuild with proper nested children
	return r.buildHierarchy(roles, nil), nil
}

func (r *rbacRepository) buildHierarchy(roles []Role, parentID *string) []RoleWithChildren {
	var result []RoleWithChildren

	for _, role := range roles {
		// Check if this role's parent matches the given parentID
		isMatch := false
		if parentID == nil && role.ParentRoleID == nil {
			isMatch = true
		} else if parentID != nil && role.ParentRoleID != nil && *parentID == *role.ParentRoleID {
			isMatch = true
		}

		if isMatch {
			children := r.buildHierarchy(roles, &role.ID)
			result = append(result, RoleWithChildren{
				Role:     role,
				Children: children,
			})
		}
	}

	return result
}

func (r *rbacRepository) GetRoleAncestors(roleID string) ([]Role, error) {
	query := `
		SELECT id, name, description, parent_role_id, COALESCE(level, 0), created_at
		FROM roles
		WHERE id IN (SELECT ancestor_id FROM get_role_ancestors($1))
		ORDER BY level DESC
	`
	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var description sql.NullString
		var parentRoleID sql.NullString
		if err := rows.Scan(&role.ID, &role.Name, &description, &parentRoleID, &role.Level, &role.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		role.Description = description.String
		if parentRoleID.Valid {
			role.ParentRoleID = &parentRoleID.String
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *rbacRepository) GetRoleDescendants(roleID string) ([]Role, error) {
	query := `
		SELECT id, name, description, parent_role_id, COALESCE(level, 0), created_at
		FROM roles
		WHERE id IN (SELECT descendant_id FROM get_role_descendants($1))
		ORDER BY level
	`
	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var description sql.NullString
		var parentRoleID sql.NullString
		if err := rows.Scan(&role.ID, &role.Name, &description, &parentRoleID, &role.Level, &role.CreatedAt); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		role.Description = description.String
		if parentRoleID.Valid {
			role.ParentRoleID = &parentRoleID.String
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *rbacRepository) GetRolePermissionsWithInheritance(roleID string) ([]InheritedPermission, error) {
	query := `
		SELECT permission_id, permission_name, resource, action, is_inherited
		FROM role_permissions_with_inheritance
		WHERE role_id = $1
		ORDER BY resource, action
	`
	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var permissions []InheritedPermission
	for rows.Next() {
		var perm InheritedPermission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.IsInherited); err != nil {
			return nil, errors.Wrap(err, errors.DatabaseScanFailed)
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (r *rbacRepository) SetParentRole(roleID, parentRoleID string) error {
	// Get parent role level
	var parentLevel int
	if parentRoleID != "" {
		parent, err := r.GetRoleByID(parentRoleID)
		if err != nil {
			return err
		}
		parentLevel = parent.Level + 1
	}

	query := `UPDATE roles SET parent_role_id = $2, level = $3 WHERE id = $1`
	var parentPtr *string
	if parentRoleID != "" {
		parentPtr = &parentRoleID
	}

	result, err := r.db.Exec(query, roleID, parentPtr, parentLevel)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}
	return nil
}
