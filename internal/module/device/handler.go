package device

import (
	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/response"
	"boilerplate-be/internal/shared/validator"

	"github.com/gofiber/fiber/v2"
)

type DeviceHandler struct {
	deviceUseCase DeviceUseCase
}

func NewDeviceHandler(deviceUseCase DeviceUseCase) *DeviceHandler {
	return &DeviceHandler{
		deviceUseCase: deviceUseCase,
	}
}

// RegisterDevice godoc
// @Summary      Register a new device
// @Description  Registers a new device and returns API key (save it, won't be shown again)
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      RegisterDeviceRequest  true  "Device data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.DeviceWithAPIKeyDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Router       /devices [post]
func (h *DeviceHandler) RegisterDevice(c *fiber.Ctx) error {
	var req RegisterDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	device, apiKey, err := h.deviceUseCase.RegisterDevice(&req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	deviceResp := DeviceWithAPIKeyResponse{
		DeviceResponse: *device.ToResponse(),
		APIKey:         apiKey,
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, deviceResp, fiber.StatusCreated,
	))
}

// GetDevice godoc
// @Summary      Get a device by ID
// @Description  Returns a device by its ID
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path      string  true  "Device UUID"
// @Success      200   {object}  docs.SuccessResponse{data=docs.DeviceDocResponse}
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /devices/{uuid} [get]
func (h *DeviceHandler) GetDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	device, err := h.deviceUseCase.GetDevice(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, device.ToResponse(),
	))
}

// GetDevices godoc
// @Summary      Get all devices
// @Description  Returns all devices with optional filters
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        room_id query     string  false  "Room ID filter"
// @Param        type    query     string  false  "Device type filter"
// @Param        status  query     string  false  "Device status filter"
// @Param        page    query     int     false  "Page number" default(1)
// @Param        limit   query     int     false  "Items per page" default(10)
// @Success      200     {object}  docs.SuccessResponse{data=[]docs.DeviceDocResponse,meta=docs.MetaResponse}
// @Failure      401     {object}  docs.ErrorResponse
// @Router       /devices [get]
func (h *DeviceHandler) GetDevices(c *fiber.Ctx) error {
	filter := &DeviceFilter{
		RoomID: c.Query("room_id"),
		Type:   DeviceType(c.Query("type")),
		Status: DeviceStatus(c.Query("status")),
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
	}

	devices, total, err := h.deviceUseCase.GetDevices(filter)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var deviceResponses []*DeviceResponse
	for _, device := range devices {
		deviceResponses = append(deviceResponses, device.ToResponse())
	}

	totalPages := int64(total / filter.Limit)
	if total%filter.Limit > 0 {
		totalPages++
	}

	meta := &response.MetaResponse{
		Page:      int64(filter.Page),
		PageSize:  int64(filter.Limit),
		Total:     int64(total),
		TotalPage: totalPages,
		IsNext:    int64(filter.Page) < totalPages,
		IsBack:    filter.Page > 1,
	}

	return c.JSON(response.CreatePaginatedResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, deviceResponses, meta,
	))
}

// UpdateDevice godoc
// @Summary      Update a device
// @Description  Updates a device by ID
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path      string              true  "Device UUID"
// @Param        body  body      UpdateDeviceRequest true  "Device update data"
// @Success      200   {object}  docs.SuccessResponse{data=docs.DeviceDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /devices/{uuid} [put]
func (h *DeviceHandler) UpdateDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	device, err := h.deviceUseCase.UpdateDevice(id, &req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataUpdated.ID, response.MsgDataUpdated.EN, device.ToResponse(),
	))
}

// UpdateDeviceStatus godoc
// @Summary      Update device status
// @Description  Updates a device status (online, offline, maintenance, error)
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path      string                    true  "Device UUID"
// @Param        body  body      UpdateDeviceStatusRequest true  "Status update"
// @Success      200   {object}  docs.SuccessResponse
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /devices/{uuid}/status [put]
func (h *DeviceHandler) UpdateDeviceStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateDeviceStatusRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.deviceUseCase.UpdateDeviceStatus(id, req.Status); err != nil {
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

// Heartbeat godoc
// @Summary      Device heartbeat
// @Description  Updates device last heartbeat timestamp
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Param        X-API-Key  header    string  true  "Device API Key"
// @Success      200        {object}  docs.ErrorResponse
// @Failure      401        {object}  docs.ErrorResponse
// @Router       /devices/heartbeat [post]
func (h *DeviceHandler) Heartbeat(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		appErr := errors.New(errors.InvalidAPIKey)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	device, err := h.deviceUseCase.ValidateAPIKey(apiKey)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InvalidAPIKey)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.deviceUseCase.Heartbeat(device.ID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Heartbeat berhasil", "Heartbeat successful", nil,
	))
}

// DeleteDevice godoc
// @Summary      Delete a device
// @Description  Deletes a device by ID
// @Tags         Devices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        uuid  path      string  true  "Device UUID"
// @Success      200   {object}  docs.SuccessResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      403   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /devices/{uuid} [delete]
func (h *DeviceHandler) DeleteDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.deviceUseCase.DeleteDevice(id); err != nil {
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
