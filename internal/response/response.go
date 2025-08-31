package response

import (
	"strings"
	"time"

	"boilerplate-be/internal/infrastructure/errors"
	"boilerplate-be/internal/infrastructure/lib"

	"github.com/gofiber/fiber/v2"
)

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

type ErrorResponseStruct struct {
	Success    bool         `json:"success"`
	Code       int          `json:"code"`        
	Message    string       `json:"message"`     
	ErrorCode  int          `json:"error_code"`  
	Errors     interface{}  `json:"errors"`      
	Timestamp  time.Time    `json:"timestamp"`
}

type SuccessResponseStruct struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`       
	Message   string      `json:"message"`    
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type PaginatedResponseStruct struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`       
	Message   string      `json:"message"`    
	Data      interface{} `json:"data"`
	Meta      interface{} `json:"meta"`
	Timestamp time.Time   `json:"timestamp"`
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

func CreateErrorResponse(c *fiber.Ctx, err errors.AppError) ErrorResponseStruct {
	lang := getLanguageFromHeader(c)

	errorResp := ErrorResponseStruct{
		Success:   false,
		Code:      err.StatusCode,
		Message:   getMessageByLanguage(err.Code.MessageID(), err.Code.MessageEN(), lang),
		ErrorCode: err.Code.Value(),
		Timestamp: time.Now(),
	}

	if validationDetails, ok := err.Details.([]lib.ValidationErrorDetailsBilingual); ok {
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
		errorResp.Errors = formattedErrors
	} else {
		errorResp.Errors = nil
	}

	return errorResp
}

func CreateSuccessResponse(c *fiber.Ctx, messageID, messageEN string, data interface{}, statusCode ...int) SuccessResponseStruct {
	lang := getLanguageFromHeader(c)
	
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	
	return SuccessResponseStruct{
		Success:   true,
		Code:      code,
		Message:   getMessageByLanguage(messageID, messageEN, lang),
		Data:      data,
		Timestamp: time.Now(),
	}
}

func CreatePaginatedResponse(c *fiber.Ctx, messageID, messageEN string, data interface{}, meta interface{}, statusCode ...int) PaginatedResponseStruct {
	lang := getLanguageFromHeader(c)
	
	code := fiber.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	
	return PaginatedResponseStruct{
		Success:   true,
		Code:      code,
		Message:   getMessageByLanguage(messageID, messageEN, lang),
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	}
}
