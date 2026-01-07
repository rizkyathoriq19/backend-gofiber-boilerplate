package patient

import (
	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"
	"boilerplate-be/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type PatientHandler struct {
	patientUseCase PatientUseCase
}

func NewPatientHandler(patientUseCase PatientUseCase) *PatientHandler {
	return &PatientHandler{
		patientUseCase: patientUseCase,
	}
}

// AdmitPatient godoc
// @Summary      Admit a new patient
// @Description  Admits a new patient to the hospital
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      AdmitPatientRequest  true  "Patient data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.PatientDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Router       /patients [post]
func (h *PatientHandler) AdmitPatient(c *fiber.Ctx) error {
	var req AdmitPatientRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	patient, err := h.patientUseCase.AdmitPatient(&req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, patient.ToResponse(), fiber.StatusCreated,
	))
}

// GetPatient godoc
// @Summary      Get a patient by ID
// @Description  Returns a patient by ID
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  docs.SuccessResponse{data=docs.PatientDocResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /patients/{id} [get]
func (h *PatientHandler) GetPatient(c *fiber.Ctx) error {
	id := c.Params("id")

	patient, err := h.patientUseCase.GetPatient(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, patient.ToResponse(),
	))
}

// GetPatients godoc
// @Summary      Get all patients
// @Description  Returns all patients with optional filters
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        room_id         query     string  false  "Room ID filter"
// @Param        condition_level query     string  false  "Condition level filter"
// @Param        is_admitted     query     bool    false  "Admission status filter"
// @Param        page            query     int     false  "Page number" default(1)
// @Param        limit           query     int     false  "Items per page" default(10)
// @Success      200             {object}  docs.SuccessResponse{data=docs.PatientListDocResponse}
// @Failure      401             {object}  docs.ErrorResponse
// @Router       /patients [get]
func (h *PatientHandler) GetPatients(c *fiber.Ctx) error {
	filter := &PatientFilter{
		RoomID:         c.Query("room_id"),
		ConditionLevel: ConditionLevel(c.Query("condition_level")),
		Page:           c.QueryInt("page", 1),
		Limit:          c.QueryInt("limit", 10),
	}

	if isAdmittedStr := c.Query("is_admitted"); isAdmittedStr != "" {
		isAdmitted := isAdmittedStr == "true"
		filter.IsAdmitted = &isAdmitted
	}

	patients, total, err := h.patientUseCase.GetAllPatients(filter)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	var patientResponses []*PatientWithRoomResponse
	for _, patient := range patients {
		patientResponses = append(patientResponses, patient.ToResponse())
	}

	totalPages := total / filter.Limit
	if total%filter.Limit > 0 {
		totalPages++
	}

	listResponse := PatientListResponse{
		Patients:   patientResponses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataRetrieved.ID, response.MsgDataRetrieved.EN, listResponse,
	))
}

// UpdatePatient godoc
// @Summary      Update a patient
// @Description  Updates a patient by ID
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string               true  "Patient ID"
// @Param        body  body      UpdatePatientRequest true  "Patient update data"
// @Success      200   {object}  docs.SuccessResponse{data=docs.PatientDocResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /patients/{id} [put]
func (h *PatientHandler) UpdatePatient(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdatePatientRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	patient, err := h.patientUseCase.UpdatePatient(id, &req)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgDataUpdated.ID, response.MsgDataUpdated.EN, patient.ToResponse(),
	))
}

// UpdateConditionLevel godoc
// @Summary      Update patient condition level
// @Description  Updates a patient's condition level
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "Patient ID"
// @Param        body  body      UpdateConditionRequest  true  "Condition update data"
// @Success      200   {object}  docs.ErrorResponse
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Router       /patients/{id}/condition [put]
func (h *PatientHandler) UpdateConditionLevel(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateConditionRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := h.patientUseCase.UpdateConditionLevel(id, req.ConditionLevel); err != nil {
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

// DischargePatient godoc
// @Summary      Discharge a patient
// @Description  Discharges a patient from the hospital
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  docs.ErrorResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /patients/{id}/discharge [post]
func (h *PatientHandler) DischargePatient(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.patientUseCase.DischargePatient(id); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, "Pasien berhasil keluar", "Patient successfully discharged", nil,
	))
}

// DeletePatient godoc
// @Summary      Delete a patient
// @Description  Deletes a patient by ID
// @Tags         Patients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Patient ID"
// @Success      200  {object}  docs.ErrorResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Failure      403  {object}  docs.ErrorResponse
// @Failure      404  {object}  docs.ErrorResponse
// @Router       /patients/{id} [delete]
func (h *PatientHandler) DeletePatient(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.patientUseCase.DeletePatient(id); err != nil {
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
