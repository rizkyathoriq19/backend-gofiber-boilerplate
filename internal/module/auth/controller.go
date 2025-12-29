package auth

import (
	"time"

	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/response"
	"boilerplate-be/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUseCase AuthUseCase
}

func NewAuthHandler(authUseCase AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register godoc
// @Summary      Register new user
// @Description  Creates a new user account and returns tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      docs.RegisterRequest  true  "Registration data"
// @Success      201   {object}  docs.SuccessResponse{data=docs.AuthResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      409   {object}  docs.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	user, accessToken, refreshToken, err := h.authUseCase.Register(req.Email, req.Password, req.Name)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	authResponse := AuthResponse{
		User: UserResponse{
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

	return c.Status(fiber.StatusCreated).JSON(response.CreateSuccessResponse(
		c, response.MsgDataCreated.ID, response.MsgDataCreated.EN, authResponse, fiber.StatusCreated,
	))
}

// Login godoc
// @Summary      User login
// @Description  Authenticates user and returns access/refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      docs.LoginRequest  true  "Login credentials"
// @Success      200   {object}  docs.SuccessResponse{data=docs.TokenResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      404   {object}  docs.ErrorResponse
// @Failure      422   {object}  docs.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		appErr := errors.New(errors.InvalidRequestBody)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	accessToken, refreshToken, err := h.authUseCase.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	tokenResponse := RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgLoginSuccess.ID, response.MsgLoginSuccess.EN, tokenResponse,
	))
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Exchanges refresh token for new access/refresh token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      docs.RefreshTokenRequest  true  "Refresh token"
// @Success      200   {object}  docs.SuccessResponse{data=docs.TokenResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.CreateErrorResponse(
			c, errors.New(errors.InvalidRequestBody),
		))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	accessToken, refreshToken, err := h.authUseCase.RefreshToken(req.RefreshToken)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	tokenResponse := RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second),
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgTokenRefresh.ID, response.MsgTokenRefresh.EN, tokenResponse,
	))
}

// Logout godoc
// @Summary      User logout
// @Description  Invalidates current access token and revokes refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse
// @Failure      401  {object}  docs.ErrorResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	tokenID := c.Locals("token_id").(string)

	if err := h.authUseCase.Logout(userID, tokenID); err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgLogoutSuccess.ID, response.MsgLogoutSuccess.EN, nil))
}

// Profile godoc
// @Summary      Get user profile
// @Description  Returns current authenticated user's profile
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  docs.SuccessResponse{data=docs.UserResponse}
// @Failure      401  {object}  docs.ErrorResponse
// @Router       /auth/profile [get]
func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.authUseCase.GetProfile(userID)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	userResponse := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgProfileRetrieve.ID, response.MsgProfileRetrieve.EN, userResponse,
	))
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates current authenticated user's profile
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      docs.UpdateProfileRequest  true  "Profile update data"
// @Success      200   {object}  docs.SuccessResponse{data=docs.UserResponse}
// @Failure      400   {object}  docs.ErrorResponse
// @Failure      401   {object}  docs.ErrorResponse
// @Router       /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.CreateErrorResponse(
			c, errors.New(errors.InvalidRequestBody),
		))
	}

	if err := validator.ValidateStruct(req); err != nil {
		validationErrors := validator.FormatValidationErrorForResponseBilingual(err)
		// Panggil error code langsung dari package 'errors'
		appErr := errors.NewWithDetails(errors.ValidationFailed, validationErrors)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	user, err := h.authUseCase.UpdateProfile(userID, req.Name)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
		}
		appErr := errors.New(errors.InternalServerError)
		return c.Status(appErr.StatusCode).JSON(response.CreateErrorResponse(c, appErr))
	}

	userResponse := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(response.CreateSuccessResponse(
		c, response.MsgProfileUpdate.ID, response.MsgProfileUpdate.EN, userResponse,
	))
}
