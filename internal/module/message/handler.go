package message

import (
	"strconv"

	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/validator"

	"github.com/gofiber/fiber/v2"
)

// MessageHandler handles message HTTP requests
type MessageHandler struct {
	messageUseCase MessageUseCase
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageUseCase MessageUseCase) *MessageHandler {
	return &MessageHandler{
		messageUseCase: messageUseCase,
	}
}

// SendMessage sends a new message
// @Summary Send a message
// @Tags Messages
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "Message request"
// @Success 201 {object} response.Response
// @Security BearerAuth
// @Router /messages [post]
func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	var req SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(&req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	// Get user info from context
	userID := c.Locals("user_id").(string)
	userRole := c.Locals("user_role").(string)
	isStaff := userRole != "patient"

	msg, err := h.messageUseCase.SendMessage(&req, userID, isStaff)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, msg, fiber.StatusCreated,
	))
}

// GetMessages gets messages for the current room
// @Summary Get messages
// @Tags Messages
// @Produce json
// @Param room_id query string true "Room ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages [get]
func (h *MessageHandler) GetMessages(c *fiber.Ctx) error {
	roomID := c.Query("room_id")
	if roomID == "" {
		appErr := errors.New(errors.ValidationFailed)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	filter := &MessageFilter{
		Page:      1,
		Limit:     20,
		Direction: c.Query("direction"),
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	messages, total, err := h.messageUseCase.GetRoomMessages(roomID, filter)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	totalPages := int64((total + filter.Limit - 1) / filter.Limit)
	meta := &response.MetaResponse{
		Page:      int64(filter.Page),
		PageSize:  int64(filter.Limit),
		Total:     int64(total),
		TotalPage: totalPages,
		IsNext:    int64(filter.Page) < totalPages,
		IsBack:    filter.Page > 1,
	}

	return c.JSON(response.CreatePaginatedResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, messages, meta,
	))
}

// GetMessage gets a message by ID
// @Summary Get a message
// @Tags Messages
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages/{id} [get]
func (h *MessageHandler) GetMessage(c *fiber.Ctx) error {
	id := c.Params("id")

	msg, err := h.messageUseCase.GetMessage(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, msg,
	))
}

// GetMyMessages gets messages for the current user
// @Summary Get my messages
// @Tags Messages
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages/my [get]
func (h *MessageHandler) GetMyMessages(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	userRole := c.Locals("user_role").(string)

	filter := &MessageFilter{
		Page:      1,
		Limit:     20,
		Direction: c.Query("direction"),
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	var messages []*MessageWithDetails
	var total int
	var err error

	if userRole == "patient" {
		messages, total, err = h.messageUseCase.GetPatientMessages(userID, filter)
	} else {
		messages, total, err = h.messageUseCase.GetStaffMessages(userID, filter)
	}

	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	totalPages := int64((total + filter.Limit - 1) / filter.Limit)
	meta := &response.MetaResponse{
		Page:      int64(filter.Page),
		PageSize:  int64(filter.Limit),
		Total:     int64(total),
		TotalPage: totalPages,
		IsNext:    int64(filter.Page) < totalPages,
		IsBack:    filter.Page > 1,
	}

	return c.JSON(response.CreatePaginatedResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, messages, meta,
	))
}

// GetUnreadCount gets unread message count
// @Summary Get unread count
// @Tags Messages
// @Produce json
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages/unread-count [get]
func (h *MessageHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	userRole := c.Locals("user_role").(string)
	isStaff := userRole != "patient"

	count, err := h.messageUseCase.GetUnreadCount(userID, isStaff)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, fiber.Map{"unread_count": count},
	))
}

// MarkAsRead marks a message as read
// @Summary Mark message as read
// @Tags Messages
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages/{id}/read [put]
func (h *MessageHandler) MarkAsRead(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.messageUseCase.MarkAsRead(id); err != nil {
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

// MarkAllAsRead marks all messages in a room as read
// @Summary Mark all messages as read
// @Tags Messages
// @Produce json
// @Param room_id query string true "Room ID"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages/read-all [put]
func (h *MessageHandler) MarkAllAsRead(c *fiber.Ctx) error {
	roomID := c.Query("room_id")
	if roomID == "" {
		appErr := errors.New(errors.ValidationFailed)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	userID := c.Locals("user_id").(string)

	if err := h.messageUseCase.MarkAllAsRead(roomID, userID); err != nil {
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

// DeleteMessage deletes a message
// @Summary Delete a message
// @Tags Messages
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} response.Response
// @Security BearerAuth
// @Router /messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.messageUseCase.DeleteMessage(id); err != nil {
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
