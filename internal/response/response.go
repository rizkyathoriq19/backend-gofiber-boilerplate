package response

import (
	"time"

	"boilerplate-be/internal/infrastructure/errors"
)

type BilingualMessage struct {
	ID string `json:"id"`
	EN string `json:"en"`
}

type ErrorDetail struct {
	Code    int         `json:"code"`
	Details interface{} `json:"details"`
}

type Response struct {
	Success   bool            `json:"success"`
	Message   BilingualMessage `json:"message"`
	Data      interface{}     `json:"data,omitempty"`
	Error     *ErrorDetail    `json:"error,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// SuccessResponse dengan message bilingual
func SuccessResponse(messageID, messageEN string, data interface{}) Response {
	return Response{
		Success: true,
		Message: BilingualMessage{
			ID: messageID,
			EN: messageEN,
		},
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ErrorResponse dengan format bilingual
func ErrorResponse(err errors.AppError) Response {
	errorDetail := &ErrorDetail{
		Code: err.Code.Value(),
	}
	
	// Handle bilingual details untuk validation errors
	if validationDetails, ok := err.Details.([]errors.ValidationErrorDetails); ok {
		errorDetail.Details = convertToBilingualValidationDetails(validationDetails)
	} else {
		// Untuk error tanpa details, set ke null
		errorDetail.Details = nil
	}

	return Response{
		Success: false,
		Message: BilingualMessage{
			ID: err.Code.MessageID(),
			EN: err.Code.MessageEN(),
		},
		Error:     errorDetail,
		Timestamp: time.Now(),
	}
}

// convertToBilingualValidationDetails mengubah ValidationErrorDetails menjadi bilingual
func convertToBilingualValidationDetails(details []errors.ValidationErrorDetails) []map[string]interface{} {
	var bilingualDetails []map[string]interface{}
	
	for _, detail := range details {
		bilingualDetail := map[string]interface{}{
			"field": detail.Field,
			"message": BilingualMessage{
				ID: detail.Message,
				EN: getEnglishValidationMessage(detail.Message),
			},
		}
		bilingualDetails = append(bilingualDetails, bilingualDetail)
	}
	
	return bilingualDetails
}

// getEnglishValidationMessage - helper function untuk translate validation messages
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

type PaginatedResponse struct {
	Success   bool            `json:"success"`
	Message   BilingualMessage `json:"message"`
	Data      interface{}     `json:"data"`
	Meta      interface{}     `json:"meta"`
	Timestamp time.Time       `json:"timestamp"`
}

func PaginatedSuccessResponse(messageID, messageEN string, data interface{}, meta interface{}) PaginatedResponse {
	return PaginatedResponse{
		Success: true,
		Message: BilingualMessage{
			ID: messageID,
			EN: messageEN,
		},
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
	}
}