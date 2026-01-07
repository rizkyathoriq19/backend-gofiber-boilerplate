package alert

import (
	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"
	"boilerplate-be/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type AlertHandler struct {
	alertUseCase AlertUseCase
}

func NewAlertHandler(alertUseCase AlertUseCase) *AlertHandler {
	return &AlertHandler{
		alertUseCase: alertUseCase,
	}
}

// CreateAlert godoc
// @Summary      Create a new alert
// @Description  Creates a new nurse call alert
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateAlertRequest  true  "Alert data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.AlertDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Router       /alerts [post]
func (h *AlertHandler) CreateAlert(c *fiber.Ctx) error {
	var req CreateAlertRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	alert, err := h.alertUseCase.CreateAlert(&req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, alert.ToResponse(), fiber.StatusCreated,
	))
}

// GetAlert godoc
// @Summary      Get an alert by ID
// @Description  Returns an alert by ID
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Alert ID"
// @Success      200  {object}  docs.SuccessResponse{data=docs.AlertDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /alerts/{id} [get]
func (h *AlertHandler) GetAlert(c *fiber.Ctx) error {
	id := c.Params("id")

	alert, err := h.alertUseCase.GetAlert(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, alert.ToResponse(),
	))
}

// GetAlerts godoc
// @Summary      Get all alerts
// @Description  Returns all alerts with optional filters
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        room_id    query     string  false  "Room ID filter"
// @Param        patient_id query     string  false  "Patient ID filter"
// @Param        type       query     string  false  "Alert type filter"
// @Param        priority   query     string  false  "Priority filter"
// @Param        status     query     string  false  "Status filter"
// @Param        page       query     int     false  "Page number" default(1)
// @Param        limit      query     int     false  "Items per page" default(10)
// @Success      200        {object}  docs.SuccessResponse{data=docs.AlertListDocResponse}
// @Failure      401        {object}  docs.ErrorResponse
// @Router       /alerts [get]
func (h *AlertHandler) GetAlerts(c *fiber.Ctx) error {
	filter := &AlertFilter{
		RoomID:    c.Query("room_id"),
		PatientID: c.Query("patient_id"),
		Type:      AlertType(c.Query("type")),
		Priority:  AlertPriority(c.Query("priority")),
		Status:    AlertStatus(c.Query("status")),
		Page:      c.QueryInt("page", 1),
		Limit:     c.QueryInt("limit", 10),
	}

	alerts, total, err := h.alertUseCase.GetAlerts(filter)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var alertResponses []*AlertWithDetailsResponse
	for _, alert := range alerts {
		alertResponses = append(alertResponses, alert.ToResponse())
	}

	totalPages := total / filter.Limit
	if total%filter.Limit > 0 {
		totalPages++
	}

	listResponse := AlertListResponse{
		Alerts:     alertResponses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, listResponse,
	))
}

// GetActiveAlerts godoc
// @Summary      Get active alerts
// @Description  Returns all active (unresolved) alerts sorted by priority
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.AlertDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Router       /alerts/active [get]
func (h *AlertHandler) GetActiveAlerts(c *fiber.Ctx) error {
	alerts, err := h.alertUseCase.GetActiveAlerts()
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var alertResponses []*AlertWithDetailsResponse
	for _, alert := range alerts {
		alertResponses = append(alertResponses, alert.ToResponse())
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, alertResponses,
	))
}

// AcknowledgeAlert godoc
// @Summary      Acknowledge an alert
// @Description  Acknowledges an alert as a staff member
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Alert ID"
// @Success      200  {object}  docs.ErrorResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /alerts/{id}/acknowledge [post]
func (h *AlertHandler) AcknowledgeAlert(c *fiber.Ctx) error {
	id := c.Params("id")
	staffID := c.Locals("staff_id")

	if staffID == nil {
		appErr := errors.New(errors.Forbidden)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	// Check if staff can handle this alert (room-based permission)
	canHandle, err := h.alertUseCase.CanStaffHandleAlert(staffID.(string), id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if !canHandle {
		appErr := errors.New(errors.Forbidden)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.alertUseCase.AcknowledgeAlert(id, staffID.(string)); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Alert berhasil di-acknowledge", "Alert acknowledged successfully", nil,
	))
}

// ResolveAlert godoc
// @Summary      Resolve an alert
// @Description  Resolves an alert
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string              true  "Alert ID"
// @Param        body  body      ResolveAlertRequest true  "Resolution notes"
// @Success      200   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /alerts/{id}/resolve [post]
func (h *AlertHandler) ResolveAlert(c *fiber.Ctx) error {
	id := c.Params("id")
	staffID := c.Locals("staff_id")

	if staffID == nil {
		appErr := errors.New(errors.Forbidden)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var req ResolveAlertRequest
	if err := c.BodyParser(&req); err != nil {
		req = ResolveAlertRequest{}
	}

	// Check room-based permission
	canHandle, err := h.alertUseCase.CanStaffHandleAlert(staffID.(string), id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if !canHandle {
		appErr := errors.New(errors.Forbidden)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.alertUseCase.ResolveAlert(id, staffID.(string), req.Notes); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Alert berhasil diselesaikan", "Alert resolved successfully", nil,
	))
}

// GetAlertHistory godoc
// @Summary      Get alert history
// @Description  Returns the history of an alert
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Alert ID"
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.AlertHistoryDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /alerts/{id}/history [get]
func (h *AlertHandler) GetAlertHistory(c *fiber.Ctx) error {
	id := c.Params("id")

	history, err := h.alertUseCase.GetAlertHistory(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var historyResponses []*AlertHistoryResponse
	for _, h := range history {
		historyResponses = append(historyResponses, &AlertHistoryResponse{
			ID:             h.ID,
			Action:         h.Action,
			PreviousStatus: h.PreviousStatus,
			NewStatus:      h.NewStatus,
			Notes:          h.Notes,
			CreatedAt:      h.CreatedAt,
		})
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, historyResponses,
	))
}
