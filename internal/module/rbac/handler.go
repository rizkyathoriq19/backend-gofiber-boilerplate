package rbac

import (
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/validator"

	"github.com/gofiber/fiber/v2"
)

type RBACHandler struct {
	rbacUseCase RBACUseCase
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(rbacUseCase RBACUseCase) *RBACHandler {
	return &RBACHandler{
		rbacUseCase: rbacUseCase,
	}
}

// ==================== Role Endpoints ====================

// GetRoles godoc
// @Summary      List all roles
// @Description  Returns all roles in the system (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.RoleResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Router       /super-admin/roles [get]
func (h *RBACHandler) GetRoles(c *fiber.Ctx) error {
	roles, err := h.rbacUseCase.GetRoles()
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Daftar role berhasil diambil", "Roles retrieved successfully", ToRoleResponses(roles),
	))
}

// GetRole godoc
// @Summary      Get role details
// @Description  Returns a role with its assigned permissions (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  docs.SuccessResponse{data=docs.RoleResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /super-admin/roles/{id} [get]
func (h *RBACHandler) GetRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	role, err := h.rbacUseCase.GetRoleByID(roleID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	// Get role permissions
	permissions, _ := h.rbacUseCase.GetRolePermissions(roleID)

	resp := RoleWithPermissionsResponse{
		RoleResponse: ToRoleResponse(role),
		Permissions:  ToPermissionResponses(permissions),
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Role berhasil diambil", "Role retrieved successfully", resp,
	))
}

// CreateRole godoc
// @Summary      Create a new role
// @Description  Creates a new role in the system (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      docs.CreateRoleRequest  true  "Role data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.RoleResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Failure      409   {object}  docs.ErrorResponse
// @Router       /super-admin/roles [post]
func (h *RBACHandler) CreateRole(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidRequestBody)))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	role, err := h.rbacUseCase.CreateRole(req.Name, req.Description)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, "Role berhasil dibuat", "Role created successfully", ToRoleResponse(role), fiber.StatusCreated,
	))
}

// UpdateRole godoc
// @Summary      Update a role
// @Description  Updates an existing role (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "Role ID"
// @Param        body  body      docs.CreateRoleRequest  true  "Role update data"
// @Success      200   {object}  docs.SuccessResponse{data=docs.RoleResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /super-admin/roles/{id} [put]
func (h *RBACHandler) UpdateRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidRequestBody)))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	role, err := h.rbacUseCase.UpdateRole(roleID, req.Name, req.Description)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Role berhasil diperbarui", "Role updated successfully", ToRoleResponse(role),
	))
}

// DeleteRole godoc
// @Summary      Delete a role
// @Description  Deletes a role (Super Admin only, cannot delete system roles)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  docs.SuccessResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /super-admin/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	if err := h.rbacUseCase.DeleteRole(roleID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Role berhasil dihapus", "Role deleted successfully", nil,
	))
}

// ==================== Permission Endpoints ====================

// GetPermissions godoc
// @Summary      List all permissions
// @Description  Returns all permissions in the system (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.PermissionResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Router       /super-admin/permissions [get]
func (h *RBACHandler) GetPermissions(c *fiber.Ctx) error {
	permissions, err := h.rbacUseCase.GetPermissions()
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Daftar permission berhasil diambil", "Permissions retrieved successfully", ToPermissionResponses(permissions),
	))
}

// GetRolePermissions godoc
// @Summary      Get role permissions
// @Description  Returns all permissions assigned to a role (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Role ID"
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.PermissionResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /super-admin/roles/{id}/permissions [get]
func (h *RBACHandler) GetRolePermissions(c *fiber.Ctx) error {
	roleID := c.Params("id")

	permissions, err := h.rbacUseCase.GetRolePermissions(roleID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Permission role berhasil diambil", "Role permissions retrieved successfully", ToPermissionResponses(permissions),
	))
}

// AssignPermissionToRole godoc
// @Summary      Assign permission to role
// @Description  Assigns a permission to a role (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                       true  "Role ID"
// @Param        body  body      docs.AssignPermissionRequest true  "Permission assignment"
// @Success      201   {object}  docs.SuccessResponse
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /super-admin/roles/{id}/permissions [post]
func (h *RBACHandler) AssignPermissionToRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	var req AssignPermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidRequestBody)))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.rbacUseCase.AssignPermissionToRole(roleID, req.PermissionID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, "Permission berhasil ditambahkan ke role", "Permission assigned to role successfully", nil, fiber.StatusCreated,
	))
}

// RemovePermissionFromRole godoc
// @Summary      Remove permission from role
// @Description  Removes a permission from a role (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id            path      string  true  "Role ID"
// @Param        permissionId  path      string  true  "Permission ID"
// @Success      200           {object}  docs.SuccessResponse
// @Failure      401           {object}  docs.ErrorResponse
// @Failure      403           {object}  docs.ErrorResponse
// @Failure      404           {object}  docs.ErrorResponse
// @Router       /super-admin/roles/{id}/permissions/{permissionId} [delete]
func (h *RBACHandler) RemovePermissionFromRole(c *fiber.Ctx) error {
	roleID := c.Params("id")
	permissionID := c.Params("permissionId")

	if err := h.rbacUseCase.RemovePermissionFromRole(roleID, permissionID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Permission berhasil dihapus dari role", "Permission removed from role successfully", nil,
	))
}

// ==================== User Role Endpoints ====================

// GetUserRoles godoc
// @Summary      Get user roles
// @Description  Returns all roles assigned to a user (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId  path      string  true  "User ID"
// @Success      200     {object}  docs.SuccessResponse{data=docs.UserRolesResponse}
// @Failure      401     {object}  docs.ErrorResponse
// @Failure      403     {object}  docs.ErrorResponse
// @Router       /super-admin/users/{userId}/roles [get]
func (h *RBACHandler) GetUserRoles(c *fiber.Ctx) error {
	userID := c.Params("userId")

	roles, err := h.rbacUseCase.GetUserRoles(userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	resp := UserRolesResponse{
		UserID: userID,
		Roles:  ToRoleResponses(roles),
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Role user berhasil diambil", "User roles retrieved successfully", resp,
	))
}

// AssignRoleToUser godoc
// @Summary      Assign role to user
// @Description  Assigns a role to a user (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId  path      string                 true  "User ID"
// @Param        body    body      docs.AssignRoleRequest true  "Role assignment"
// @Success      201     {object}  docs.SuccessResponse
// @Failure      400     {object}  docs.ErrorResponse
// @Failure      401     {object}  docs.ErrorResponse
// @Failure      403     {object}  docs.ErrorResponse
// @Failure      404     {object}  docs.ErrorResponse
// @Router       /super-admin/users/{userId}/roles [post]
func (h *RBACHandler) AssignRoleToUser(c *fiber.Ctx) error {
	userID := c.Params("userId")

	var req AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.CreateErrorResponse(c, errors.New(errors.InvalidRequestBody)))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.rbacUseCase.AssignRoleToUser(userID, req.RoleID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, "Role berhasil ditambahkan ke user", "Role assigned to user successfully", nil, fiber.StatusCreated,
	))
}

// RemoveRoleFromUser godoc
// @Summary      Remove role from user
// @Description  Removes a role from a user (Super Admin only)
// @Tags         Super Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        userId  path      string  true  "User ID"
// @Param        roleId  path      string  true  "Role ID"
// @Success      200     {object}  docs.SuccessResponse
// @Failure      401     {object}  docs.ErrorResponse
// @Failure      403     {object}  docs.ErrorResponse
// @Router       /super-admin/users/{userId}/roles/{roleId} [delete]
func (h *RBACHandler) RemoveRoleFromUser(c *fiber.Ctx) error {
	userID := c.Params("userId")
	roleID := c.Params("roleId")

	if err := h.rbacUseCase.RemoveRoleFromUser(userID, roleID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Role berhasil dihapus dari user", "Role removed from user successfully", nil,
	))
}

// GetMyRoles godoc
// @Summary      Get my roles
// @Description  Returns roles for the current authenticated user
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse{data=docs.UserRolesResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Router       /auth/my-roles [get]
func (h *RBACHandler) GetMyRoles(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	roles, err := h.rbacUseCase.GetUserRoles(userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	resp := UserRolesResponse{
		UserID: userID,
		Roles:  ToRoleResponses(roles),
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Role Anda berhasil diambil", "Your roles retrieved successfully", resp,
	))
}

// GetMyPermissions godoc
// @Summary      Get my permissions
// @Description  Returns permissions for the current authenticated user
// @Tags         RBAC
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.PermissionResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Router       /auth/my-permissions [get]
func (h *RBACHandler) GetMyPermissions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	permissions, err := h.rbacUseCase.GetUserPermissions(userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.CreateErrorResponse(c, errors.New(errors.InternalServerError)))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Permission Anda berhasil diambil", "Your permissions retrieved successfully", ToPermissionResponses(permissions),
	))
}
