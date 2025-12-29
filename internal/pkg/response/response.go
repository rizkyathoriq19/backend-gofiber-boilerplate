package response

import (
	"strings"
	"time"

	"boilerplate-be/internal/pkg/errors"
	"boilerplate-be/internal/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type BaseResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	ErrorCode int         `json:"error_code,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Meta      *MetaResponse `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type MetaResponse struct {
	Page      int64 `json:"page"`
	PageSize  int64 `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"total_page"`
	IsNext    bool  `json:"is_next"`
	IsBack    bool  `json:"is_back"`
}

type BilingualMessage struct {
	ID string `json:"id"`
	EN string `json:"en"`
}

type ValidationError struct {
	Field   string           `json:"field"`
	Message BilingualMessage `json:"message"`
}

type FormattedValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e BaseResponse) Error() string {
	panic("unimplemented")
}

func getLanguageFromHeader(c *fiber.Ctx) string {
	acceptLanguage := c.Get("Accept-Language", "id")
	if strings.Contains(acceptLanguage, "en") {
		return "en"
	}
	return "id"
}

func getMessageByLanguage(messageID, messageEN, language string) string {
	if language == "en" {
		return messageEN
	}
	return messageID
}

func CreateErrorResponse(c *fiber.Ctx, err errors.AppError) BaseResponse {
	lang := getLanguageFromHeader(c)

	resp := BaseResponse{
		Success:   false,
		Code:      err.StatusCode,
		Message:   getMessageByLanguage(err.Code.MessageID(), err.Code.MessageEN(), lang),
		ErrorCode: err.Code.Value(),
		Timestamp: time.Now(),
	}

	if validationDetails, ok := err.Details.([]validator.ValidationErrorDetailsBilingual); ok {
		var formattedErrors []FormattedValidationError
		for _, detail := range validationDetails {
			message := detail.Message.ID
			if lang == "en" {
				message = detail.Message.EN
			}
			formattedErrors = append(formattedErrors, FormattedValidationError{
				Field:   detail.Field,
				Message: message,
			})
		}
		resp.Errors = formattedErrors
	}

	return resp
}

func CreateSuccessResponse(c *fiber.Ctx, messageID, messageEN string, data interface{}, statusCode ...int) BaseResponse {
	lang := getLanguageFromHeader(c)
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	return BaseResponse{
		Success:   true,
		Code:      code,
		Message:   getMessageByLanguage(messageID, messageEN, lang),
		Data:      data,
		Timestamp: time.Now(),
	}
}

func CreatePaginatedResponse(c *fiber.Ctx, messageID, messageEN string, data interface{}, meta *MetaResponse, statusCode ...int) BaseResponse {
	lang := getLanguageFromHeader(c)
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	return BaseResponse{
		Success:   true,
		Code:      code,
		Message:   getMessageByLanguage(messageID, messageEN, lang),
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	}
}
