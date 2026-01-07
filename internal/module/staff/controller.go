package staff

import (
	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"
	"boilerplate-be/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type StaffHandler struct {
	staffUseCase StaffUseCase
}

func NewStaffHandler(staffUseCase StaffUseCase) *StaffHandler {
	return &StaffHandler{
		staffUseCase: staffUseCase,
	}
}

// CreateStaff godoc
// @Summary      Create a new staff member
// @Description  Creates a new staff profile for a user
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateStaffRequest  true  "Staff data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.StaffDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Router       /staff [post]
func (h *StaffHandler) CreateStaff(c *fiber.Ctx) error {
	var req CreateStaffRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	staff, err := h.staffUseCase.CreateStaff(&req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, staff.ToResponse(), fiber.StatusCreated,
	))
}

// GetStaff godoc
// @Summary      Get a staff member by ID
// @Description  Returns a staff member by ID
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Staff ID"
// @Success      200  {object}  docs.SuccessResponse{data=docs.StaffDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /staff/{id} [get]
func (h *StaffHandler) GetStaff(c *fiber.Ctx) error {
	id := c.Params("id")

	staff, err := h.staffUseCase.GetStaff(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, staff.ToResponse(),
	))
}

// GetAllStaff godoc
// @Summary      Get all staff members
// @Description  Returns all staff members with optional filters
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        type       query     string  false  "Staff type filter"
// @Param        department query     string  false  "Department filter"
// @Param        shift      query     string  false  "Shift filter"
// @Param        on_duty    query     bool    false  "On duty filter"
// @Param        page       query     int     false  "Page number" default(1)
// @Param        limit      query     int     false  "Items per page" default(10)
// @Success      200        {object}  docs.SuccessResponse{data=docs.StaffListDocResponse}
// @Failure      401        {object}  docs.ErrorResponse
// @Router       /staff [get]
func (h *StaffHandler) GetAllStaff(c *fiber.Ctx) error {
	filter := &StaffFilter{
		Type:       StaffType(c.Query("type")),
		Department: c.Query("department"),
		Shift:      ShiftType(c.Query("shift")),
		Page:       c.QueryInt("page", 1),
		Limit:      c.QueryInt("limit", 10),
	}

	if onDutyStr := c.Query("on_duty"); onDutyStr != "" {
		onDuty := onDutyStr == "true"
		filter.OnDuty = &onDuty
	}

	staffList, total, err := h.staffUseCase.GetAllStaff(filter)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var staffResponses []*StaffWithUserResponse
	for _, staff := range staffList {
		staffResponses = append(staffResponses, staff.ToResponse())
	}

	totalPages := total / filter.Limit
	if total%filter.Limit > 0 {
		totalPages++
	}

	listResponse := StaffListResponse{
		Staff:      staffResponses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, listResponse,
	))
}

// GetOnDutyStaff godoc
// @Summary      Get on-duty staff
// @Description  Returns all on-duty staff, optionally filtered by type
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        type  query     string  false  "Staff type filter (nurse, doctor)"
// @Success      200   {object}  docs.SuccessResponse{data=[]docs.StaffDocResponse}
// @Failure      401   {object}  docs.ErrorResponse
// @Router       /staff/on-duty [get]
func (h *StaffHandler) GetOnDutyStaff(c *fiber.Ctx) error {
	staffType := StaffType(c.Query("type"))

	var staffList []*StaffWithUser
	var err error

	if staffType != "" {
		staffList, err = h.staffUseCase.GetOnDutyStaff(staffType)
	} else {
		// Get all on-duty staff
		filter := &StaffFilter{}
		onDuty := true
		filter.OnDuty = &onDuty
		filter.Limit = 100
		staffList, _, err = h.staffUseCase.GetAllStaff(filter)
	}

	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var staffResponses []*StaffWithUserResponse
	for _, staff := range staffList {
		staffResponses = append(staffResponses, staff.ToResponse())
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, staffResponses,
	))
}

// UpdateStaff godoc
// @Summary      Update a staff member
// @Description  Updates a staff member by ID
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string             true  "Staff ID"
// @Param        body  body      UpdateStaffRequest true  "Staff update data"
// @Success      200   {object}  docs.SuccessResponse{data=docs.StaffDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /staff/{id} [put]
func (h *StaffHandler) UpdateStaff(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStaffRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	staff, err := h.staffUseCase.UpdateStaff(id, &req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataUpdated.ID, response.MsgDataUpdated.EN, staff.ToResponse(),
	))
}

// UpdateShift godoc
// @Summary      Update staff shift
// @Description  Updates a staff member's shift
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string             true  "Staff ID"
// @Param        body  body      UpdateShiftRequest true  "Shift update data"
// @Success      200   {object}  docs.ErrorResponse
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /staff/{id}/shift [put]
func (h *StaffHandler) UpdateShift(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateShiftRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.staffUseCase.UpdateShift(id, req.Shift); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataUpdated.ID, response.MsgDataUpdated.EN, nil,
	))
}

// ToggleOnDuty godoc
// @Summary      Toggle on-duty status
// @Description  Toggles the on-duty status of a staff member
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Staff ID"
// @Success      200  {object}  docs.ErrorResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /staff/{id}/toggle-duty [post]
func (h *StaffHandler) ToggleOnDuty(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.staffUseCase.ToggleOnDuty(id); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataUpdated.ID, response.MsgDataUpdated.EN, nil,
	))
}

// AssignToRoom godoc
// @Summary      Assign staff to room
// @Description  Assigns a staff member to a room
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string            true  "Staff ID"
// @Param        body  body      AssignRoomRequest true  "Room assignment data"
// @Success      200   {object}  docs.ErrorResponse
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /staff/{id}/rooms [post]
func (h *StaffHandler) AssignToRoom(c *fiber.Ctx) error {
	staffID := c.Params("id")

	var req AssignRoomRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.staffUseCase.AssignToRoom(staffID, req.RoomID, req.IsPrimary); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Staff berhasil di-assign ke ruangan", "Staff successfully assigned to room", nil,
	))
}

// RemoveFromRoom godoc
// @Summary      Remove staff from room
// @Description  Removes a staff member from a room assignment
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      string  true  "Staff ID"
// @Param        roomId  path      string  true  "Room ID"
// @Success      200     {object}  docs.ErrorResponse
// @Failure      401     {object}  docs.ErrorResponse
// @Failure      404     {object}  docs.ErrorResponse
// @Router       /staff/{id}/rooms/{roomId} [delete]
func (h *StaffHandler) RemoveFromRoom(c *fiber.Ctx) error {
	staffID := c.Params("id")
	roomID := c.Params("roomId")

	if err := h.staffUseCase.RemoveFromRoom(staffID, roomID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Staff berhasil dihapus dari ruangan", "Staff successfully removed from room", nil,
	))
}

// GetRoomAssignments godoc
// @Summary      Get staff room assignments
// @Description  Returns all room assignments for a staff member
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Staff ID"
// @Success      200  {object}  docs.SuccessResponse{data=[]docs.RoomAssignmentDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /staff/{id}/rooms [get]
func (h *StaffHandler) GetRoomAssignments(c *fiber.Ctx) error {
	staffID := c.Params("id")

	assignments, err := h.staffUseCase.GetRoomAssignments(staffID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var assignmentResponses []*RoomAssignmentResponse
	for _, a := range assignments {
		assignmentResponses = append(assignmentResponses, &RoomAssignmentResponse{
			ID:         a.ID,
			RoomID:     a.RoomID,
			IsPrimary:  a.IsPrimary,
			AssignedAt: a.AssignedAt,
		})
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, assignmentResponses,
	))
}

// DeleteStaff godoc
// @Summary      Delete a staff member
// @Description  Deletes a staff member by ID
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Staff ID"
// @Success      200  {object}  docs.ErrorResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /staff/{id} [delete]
func (h *StaffHandler) DeleteStaff(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.staffUseCase.DeleteStaff(id); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataDeleted.ID, response.MsgDataDeleted.EN, nil,
	))
}
