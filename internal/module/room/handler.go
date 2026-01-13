package room

import (
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/validator"

	"github.com/gofiber/fiber/v2"
)

type RoomHandler struct {
	roomUseCase RoomUseCase
}

func NewRoomHandler(roomUseCase RoomUseCase) *RoomHandler {
	return &RoomHandler{
		roomUseCase: roomUseCase,
	}
}

// CreateRoom godoc
// @Summary      Create a new room
// @Description  Creates a new room in the hospital
// @Tags         Rooms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      CreateRoomRequest  true  "Room data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.RoomDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Router       /rooms [post]
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	var req CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	room, err := h.roomUseCase.CreateRoom(&req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, room.ToResponse(), fiber.StatusCreated,
	))
}

// GetRoom godoc
// @Summary      Get a room by ID
// @Description  Returns a room by its ID
// @Tags         Rooms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID"
// @Success      200  {object}  docs.SuccessResponse{data=docs.RoomDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /rooms/{id} [get]
func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	id := c.Params("id")

	room, err := h.roomUseCase.GetRoom(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, room.ToResponse(),
	))
}

// GetRooms godoc
// @Summary      Get all rooms
// @Description  Returns all rooms with optional filters and pagination
// @Tags         Rooms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        type      query     string  false  "Room type filter"
// @Param        floor     query     string  false  "Floor filter"
// @Param        building  query     string  false  "Building filter"
// @Param        is_active query     bool    false  "Active status filter"
// @Param        page      query     int     false  "Page number" default(1)
// @Param        limit     query     int     false  "Items per page" default(10)
// @Success      200       {object}  docs.SuccessResponse{data=docs.RoomListDocResponse}
// @Failure      401       {object}  docs.ErrorResponse
// @Router       /rooms [get]
func (h *RoomHandler) GetRooms(c *fiber.Ctx) error {
	filter := &RoomFilter{
		Type:     RoomType(c.Query("type")),
		Floor:    c.Query("floor"),
		Building: c.Query("building"),
		Page:     c.QueryInt("page", 1),
		Limit:    c.QueryInt("limit", 10),
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive := isActiveStr == "true"
		filter.IsActive = &isActive
	}

	rooms, total, err := h.roomUseCase.GetRooms(filter)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var roomResponses []*RoomResponse
	for _, room := range rooms {
		roomResponses = append(roomResponses, room.ToResponse())
	}

	totalPages := total / filter.Limit
	if total%filter.Limit > 0 {
		totalPages++
	}

	listResponse := RoomListResponse{
		Rooms:      roomResponses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, listResponse,
	))
}

// UpdateRoom godoc
// @Summary      Update a room
// @Description  Updates a room by ID
// @Tags         Rooms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string             true  "Room ID"
// @Param        body  body      UpdateRoomRequest  true  "Room update data"
// @Success      200   {object}  docs.SuccessResponse{data=docs.RoomDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /rooms/{id} [put]
func (h *RoomHandler) UpdateRoom(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	room, err := h.roomUseCase.UpdateRoom(id, &req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataUpdated.ID, response.MsgDataUpdated.EN, room.ToResponse(),
	))
}

// DeleteRoom godoc
// @Summary      Delete a room
// @Description  Deletes a room by ID
// @Tags         Rooms
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Room ID"
// @Success      200  {object}  docs.SuccessResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /rooms/{id} [delete]
func (h *RoomHandler) DeleteRoom(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.roomUseCase.DeleteRoom(id); err != nil {
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
