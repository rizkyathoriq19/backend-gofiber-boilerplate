package audit

import (
	"context"

	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type AuditHandler struct {
	useCase AuditUseCase
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(useCase AuditUseCase) *AuditHandler {
	return &AuditHandler{useCase: useCase}
}

// List godoc
// @Summary List audit logs
// @Description Get a paginated list of audit logs with optional filters
// @Tags Audit
// @Accept json
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param action query string false "Filter by action (CREATE, READ, UPDATE, DELETE, LOGIN, LOGOUT)"
// @Param resource_type query string false "Filter by resource type"
// @Param resource_id query string false "Filter by resource ID"
// @Param start_date query string false "Filter by start date (RFC3339)"
// @Param end_date query string false "Filter by end date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Security BearerAuth
// @Router /super-admin/audit-logs [get]
func (h *AuditHandler) List(c *fiber.Ctx) error {
	var req AuditLogListRequest
	if err := c.QueryParser(&req); err != nil {
		return errors.New(errors.ValidationFailed)
	}

	filter := req.ToFilter()
	logs, total, err := h.useCase.List(context.Background(), filter)
	if err != nil {
		return err
	}

	// Calculate pagination meta
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)

	return c.Status(fiber.StatusOK).JSON(response.CreatePaginatedResponse(
		c,
		"Berhasil mengambil audit logs",
		"Successfully retrieved audit logs",
		ToResponseList(logs),
		&response.MetaResponse{
			Page:      int64(page),
			PageSize:  int64(pageSize),
			Total:     total,
			TotalPage: totalPage,
			IsNext:    int64(page) < totalPage,
			IsBack:    page > 1,
		},
	))
}

// GetByID godoc
// @Summary Get audit log by ID
// @Description Get a specific audit log entry by its ID
// @Tags Audit
// @Accept json
// @Produce json
// @Param id path string true "Audit Log ID"
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Security BearerAuth
// @Router /super-admin/audit-logs/{id} [get]
func (h *AuditHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return errors.New(errors.ValidationFailed)
	}

	log, err := h.useCase.GetByID(context.Background(), id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.CreateSuccessResponse(
		c,
		"Berhasil mengambil audit log",
		"Successfully retrieved audit log",
		log.ToResponse(),
	))
}

// Cleanup godoc
// @Summary Cleanup old audit logs
// @Description Delete audit logs older than specified days (minimum 30 days)
// @Tags Audit
// @Accept json
// @Produce json
// @Param days query int false "Retention days (minimum 30)" default(90)
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Security BearerAuth
// @Router /super-admin/audit-logs/cleanup [post]
func (h *AuditHandler) Cleanup(c *fiber.Ctx) error {
	days := c.QueryInt("days", 90)

	deleted, err := h.useCase.Cleanup(context.Background(), days)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(response.CreateSuccessResponse(
		c,
		"Berhasil membersihkan audit logs lama",
		"Successfully cleaned up old audit logs",
		fiber.Map{
			"deleted_count": deleted,
		},
	))
}
