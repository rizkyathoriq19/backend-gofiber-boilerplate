package lib

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

type ValidationMessageConfig struct {
	ID map[string]string
	EN map[string]string
}

var MessageConfig = &ValidationMessageConfig{
	ID: map[string]string{
		"required":  "%s harus diisi",
		"email":     "%s harus berupa email yang valid", 
		"min":       "%s minimal %s karakter",
		"max":       "%s maksimal %s karakter",
		"eqfield":   "%s harus sama dengan %s",
		"oneof":     "%s harus salah satu dari: %s",
		"numeric":   "%s harus berupa angka",
		"alpha":     "%s harus berupa huruf",
		"alphanum":  "%s harus berupa huruf atau angka",
		"len":       "%s harus tepat %s karakter",
		"gte":       "%s harus lebih besar atau sama dengan %s",
		"lte":       "%s harus lebih kecil atau sama dengan %s",
		"gt":        "%s harus lebih besar dari %s",
		"lt":        "%s harus lebih kecil dari %s",
		"url":       "%s harus berupa URL yang valid",
		"uuid":      "%s harus berupa UUID yang valid",
		"datetime":  "%s harus berupa format tanggal yang valid",
		"default":   "%s tidak valid",
	},
	EN: map[string]string{
		"required":  "%s is required",
		"email":     "%s must be a valid email",
		"min":       "%s must be at least %s characters",
		"max":       "%s must not exceed %s characters", 
		"eqfield":   "%s must be equal to %s",
		"oneof":     "%s must be one of: %s",
		"numeric":   "%s must be numeric",
		"alpha":     "%s must contain only letters",
		"alphanum":  "%s must contain only letters and numbers",
		"len":       "%s must be exactly %s characters",
		"gte":       "%s must be greater than or equal to %s",
		"lte":       "%s must be less than or equal to %s",
		"gt":        "%s must be greater than %s",
		"lt":        "%s must be less than %s",
		"url":       "%s must be a valid URL",
		"uuid":      "%s must be a valid UUID",
		"datetime":  "%s must be a valid datetime format",
		"default":   "%s is invalid",
	},
}

var CustomMessages = map[string]map[string]string{
	"ID": {
		// Format: "field.tag": "custom message"
		// Example: "password.min": "Password harus minimal 8 karakter untuk keamanan"
	},
	"EN": {
		// Format: "field.tag": "custom message"  
		// Example: "password.min": "Password must be at least 8 characters for security"
	},
}

func init() {
	validate = validator.New()
	
	// Register custom tag name function
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type BilingualValidationError struct {
	Field   string `json:"field"`
	Message struct {
		ID string `json:"id"`
		EN string `json:"en"`
	} `json:"message"`
}

type ValidationErrorDetailsBilingual struct {
	Field   string `json:"field"`
	Message struct {
		ID string `json:"id"`
		EN string `json:"en"`
	} `json:"message"`
}

func FormatValidationErrorForResponseBilingual(err error) []ValidationErrorDetailsBilingual {
	var errors []ValidationErrorDetailsBilingual
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()
			
			validationError := ValidationErrorDetailsBilingual{
				Field: field,
			}
			validationError.Message.ID = GetValidationMessage(field, tag, param, "ID")
			validationError.Message.EN = GetValidationMessage(field, tag, param, "EN")
			
			errors = append(errors, validationError)
		}
	}
	
	return errors
}

type ValidationErrorDetails struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationError(err error) []ValidationErrorDetail {
	var errors []ValidationErrorDetail
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			
			message := GetValidationMessage(field, tag, e.Param(), "ID")
			
			errors = append(errors, ValidationErrorDetail{
				Field:   field,
				Message: message,
			})
		}
	}
	
	return errors
}

func FormatValidationErrorBilingual(err error) []BilingualValidationError {
	var errors []BilingualValidationError
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			
			messageID := GetValidationMessage(field, tag, e.Param(), "ID")
			messageEN := GetValidationMessage(field, tag, e.Param(), "EN")
			
			validationError := BilingualValidationError{
				Field: field,
			}
			validationError.Message.ID = messageID
			validationError.Message.EN = messageEN
			
			errors = append(errors, validationError)
		}
	}
	
	return errors
}

func FormatValidationErrorForResponse(err error) []ValidationErrorDetails {
	var errors []ValidationErrorDetails
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			
			message := GetValidationMessage(field, tag, e.Param(), "ID")
			
			errors = append(errors, ValidationErrorDetails{
				Field:   field,
				Message: message,
			})
		}
	}
	
	return errors
}

func GetValidationMessage(field, tag, param, language string) string {
	customKey := strings.ToLower(field) + "." + tag
	if customMsg, exists := CustomMessages[language][customKey]; exists {
		if strings.Contains(customMsg, "%s") {
			// Handle parameters in custom message
			switch tag {
			case "min", "max", "len", "gte", "lte", "gt", "lt":
				return fmt.Sprintf(customMsg, field, param)
			case "eqfield":
				return fmt.Sprintf(customMsg, field, param)
			case "oneof":
				return fmt.Sprintf(customMsg, field, strings.Replace(param, " ", ", ", -1))
			default:
				return fmt.Sprintf(customMsg, field)
			}
		}
		return customMsg
	}
	
	// Use default message templates
	var messageMap map[string]string
	if language == "EN" {
		messageMap = MessageConfig.EN
	} else {
		messageMap = MessageConfig.ID
	}
	
	template, exists := messageMap[tag]
	if !exists {
		template = messageMap["default"]
	}
	
	// Format message based on tag type
	switch tag {
	case "min", "max", "len", "gte", "lte", "gt", "lt":
		return fmt.Sprintf(template, field, param)
	case "eqfield":
		return fmt.Sprintf(template, field, param)
	case "oneof":
		return fmt.Sprintf(template, field, strings.Replace(param, " ", ", ", -1))
	default:
		return fmt.Sprintf(template, field)
	}
}

func SetCustomMessage(field, tag, messageID, messageEN string) {
	key := strings.ToLower(field) + "." + tag
	CustomMessages["ID"][key] = messageID
	CustomMessages["EN"][key] = messageEN
}

func SetMessageTemplate(tag, messageID, messageEN string) {
	MessageConfig.ID[tag] = messageID
	MessageConfig.EN[tag] = messageEN
}

func AddCustomValidationRule(tag string, messageID, messageEN string, validationFunc validator.Func) error {
	if err := validate.RegisterValidation(tag, validationFunc); err != nil {
		return err
	}
	
	MessageConfig.ID[tag] = messageID
	MessageConfig.EN[tag] = messageEN
	
	return nil
}

func GetValidationMessageEN(field, tag, param string) string {
	return GetValidationMessage(field, tag, param, "EN")
}

func GetValidationMessageID(field, tag, param string) string {
	return GetValidationMessage(field, tag, param, "ID")
}