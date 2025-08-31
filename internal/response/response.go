package response

import (
	"strings"
	"time"

	"boilerplate-be/internal/infrastructure/errors"

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
	
	if strings.Contains(acceptLanguage, "id") {
		return "id"
	}
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
		Success:    false,
		Code:       err.StatusCode,
		Message:    getMessageByLanguage(err.Code.MessageID(), err.Code.MessageEN(), lang),
		ErrorCode:  err.Code.Value(),
		Timestamp:  time.Now(),
	}
	
	// Handle validation errors
	if validationDetails, ok := err.Details.([]errors.ValidationErrorDetails); ok {
		errorResp.Errors = convertToValidationErrors(validationDetails, lang)
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

func convertToValidationErrors(details []errors.ValidationErrorDetails, _ string) []ValidationError {
	var validationErrors []ValidationError
	
	for _, detail := range details {
		validationError := ValidationError{
			Field: detail.Field,
			Message: BilingualMessage{
				ID: detail.Message,
				EN: getEnglishValidationMessage(detail.Message),
			},
		}
		validationErrors = append(validationErrors, validationError)
	}
	
	return validationErrors
}

func getEnglishValidationMessage(idMessage string) string {
	messageMap := map[string]string{
		"Format email tidak valid":          "Invalid email format",
		"Password minimal 8 karakter":       "Password must be at least 8 characters",
		"Field harus diisi":                 "Field is required",
		"Data tidak valid":                  "Invalid data",
		"Format tidak sesuai":               "Invalid format",
		"Harus berupa angka":                "Must be a number",
		"Harus berupa teks":                 "Must be text",
		"Nilainya terlalu pendek":           "Value is too short",
		"Nilainya terlalu panjang":          "Value is too long",
		"Tidak sesuai dengan konfirmasi":    "Does not match confirmation",
		"Email sudah digunakan":             "Email has already been used",
		"Nama pengguna sudah digunakan":     "Username has already been used",
	}
	
	if enMessage, exists := messageMap[idMessage]; exists {
		return enMessage
	}
	
	return idMessage
}