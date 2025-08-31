package handler

import (
	"time"

	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/lib"
	"boilerplate-be/internal/module/auth/domain"
	"boilerplate-be/internal/module/auth/dto"
	"boilerplate-be/internal/response"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUseCase domain.AuthUseCase
}

func NewAuthHandler(authUseCase domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.New(errors.BodyRequestError, "en"), // Language parameter masih diperlukan untuk AppError internal
		))
	}

	if err := lib.ValidateStruct(req); err != nil {
		validationErrors := lib.FormatValidationError(err)
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.NewWithDetails(errors.ValidationError, validationErrors, "en"),
		))
	}

	user, accessToken, refreshToken, err := h.authUseCase.Register(req.Email, req.Password, req.Name)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(
			errors.New(errors.ServerError, "en"),
		))
	}

	authResponse := dto.AuthResponse{
		User: dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	}

	return c.Status(fiber.StatusCreated).JSON(response.SuccessResponse(
		response.MsgRegisterSuccess.ID,
		response.MsgRegisterSuccess.EN,
		authResponse,
	))
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.New(errors.BodyRequestError, "en"),
		))
	}

	if err := lib.ValidateStruct(req); err != nil {
		validationErrors := lib.FormatValidationError(err)
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.NewWithDetails(errors.ValidationError, validationErrors, "en"),
		))
	}

	accessToken, refreshToken, err := h.authUseCase.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(
			errors.New(errors.ServerError, "en"),
		))
	}

	tokenResponse := dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	}

	return c.JSON(response.SuccessResponse(
		response.MsgLoginSuccess.ID,
		response.MsgLoginSuccess.EN,
		tokenResponse,
	))
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.New(errors.BodyRequestError, "en"),
		))
	}

	if err := lib.ValidateStruct(req); err != nil {
		validationErrors := lib.FormatValidationError(err)
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.NewWithDetails(errors.ValidationError, validationErrors, "en"),
		))
	}

	accessToken, refreshToken, err := h.authUseCase.RefreshToken(req.RefreshToken)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(
			errors.New(errors.ServerError, "en"),
		))
	}

	tokenResponse := dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	}

	return c.JSON(response.SuccessResponse(
		response.MsgTokenRefresh.ID,
		response.MsgTokenRefresh.EN,
		tokenResponse,
	))
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	tokenID := c.Locals("token_id").(string)

	if err := h.authUseCase.Logout(userID, tokenID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(
			errors.New(errors.ServerError, "en"),
		))
	}

	return c.JSON(response.SuccessResponse(
		response.MsgLogoutSuccess.ID,
		response.MsgLogoutSuccess.EN,
		nil,
	))
}

func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.authUseCase.GetProfile(userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(
			errors.New(errors.ServerError, "en"),
		))
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(response.SuccessResponse(
		response.MsgProfileRetrieve.ID,
		response.MsgProfileRetrieve.EN,
		userResponse,
	))
}

func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.New(errors.BodyRequestError, "en"),
		))
	}

	if err := lib.ValidateStruct(req); err != nil {
		validationErrors := lib.FormatValidationError(err)
		return c.Status(fiber.StatusBadRequest).JSON(response.ErrorResponse(
			errors.NewWithDetails(errors.ValidationError, validationErrors, "en"),
		))
	}

	user, err := h.authUseCase.UpdateProfile(userID, req.Name)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.ErrorResponse(appErr))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponse(
			errors.New(errors.ServerError, "en"),
		))
	}

	userResponse := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(response.SuccessResponse(
		response.MsgProfileUpdate.ID,
		response.MsgProfileUpdate.EN,
		userResponse,
	))
}