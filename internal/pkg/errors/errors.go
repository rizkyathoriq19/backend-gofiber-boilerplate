package errors

import (
	"boilerplate-be/internal/pkg/enum"
	"errors"
	"fmt"
)

const (
	// General Success/Info
	Success    = enum.Success
	NoDataFound = enum.NoDataFound
	DataNotFound = enum.DataNotFound

	// Client Errors
	InvalidRequest       = enum.InvalidRequest
	InvalidRequestBody   = enum.InvalidRequestBody
	MissingRequiredField = enum.MissingRequiredField
	InvalidFormat        = enum.InvalidFormat
	InvalidCredentials   = enum.InvalidCredentials
	Unauthorized         = enum.Unauthorized
	Forbidden            = enum.Forbidden
	ResourceNotFound     = enum.ResourceNotFound
	Conflict             = enum.Conflict
	ValidationFailed     = enum.ValidationFailed
	InvalidToken         = enum.InvalidToken
	TokenExpired         = enum.TokenExpired
	RateLimitExceeded    = enum.RateLimitExceeded
	InvalidAPIKey        = enum.InvalidAPIKey
	ResourceAlreadyExists = enum.ResourceAlreadyExists

	// User/Account Errors
	UsernameExists     = enum.UsernameExists
	EmailExists        = enum.EmailExists
	InvalidUsername    = enum.InvalidUsername
	InvalidEmail       = enum.InvalidEmail
	AccountNotFound    = enum.AccountNotFound
	AccountInactive    = enum.AccountInactive
	PasswordMismatch   = enum.PasswordMismatch
	AccountLocked      = enum.AccountLocked
	AccountNotVerified = enum.AccountNotVerified
	PasswordTooWeak    = enum.PasswordTooWeak

	// File Handling Errors
	FileSizeExceeded = enum.FileSizeExceeded
	InvalidFileType  = enum.InvalidFileType

	// Server Errors
	InternalServerError  = enum.InternalServerError
	DatabaseError        = enum.DatabaseError
	CacheError           = enum.CacheError
	ExternalServiceError = enum.ExternalServiceError
	ConfigurationError   = enum.ConfigurationError
	ServiceUnavailable   = enum.ServiceUnavailable

	// Database Specific
	DatabaseConnectionFailed = enum.DatabaseConnectionFailed
	DatabaseQueryFailed      = enum.DatabaseQueryFailed
	DatabaseInsertFailed     = enum.DatabaseInsertFailed
	DatabaseUpdateFailed     = enum.DatabaseUpdateFailed
	DatabaseDeleteFailed     = enum.DatabaseDeleteFailed
	DatabaseScanFailed       = enum.DatabaseScanFailed
	ForeignKeyViolation      = enum.ForeignKeyViolation
	TransactionFailed        = enum.TransactionFailed

	// Cache/Redis Specific
	CacheConnectionFailed = enum.CacheConnectionFailed
	CacheStoreFailed      = enum.CacheStoreFailed
	CacheRetrieveFailed   = enum.CacheRetrieveFailed
	CacheDeleteFailed     = enum.CacheDeleteFailed

	// Authentication Service
	TokenGenerationFailed  = enum.TokenGenerationFailed
	PasswordHashFailed     = enum.PasswordHashFailed
	AuthServiceUnavailable = enum.AuthServiceUnavailable

	// File Storage Service
	FileStorageError = enum.FileStorageError
)

type AppError struct {
	Code       enum.ErrorCode `json:"code"`
	Message    string         `json:"message"`
	StatusCode int            `json:"-"`
	Details    interface{}    `json:"details,omitempty"`
	Err        error          `json:"-"`
	Language   string         `json:"-"`
}

func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e AppError) Unwrap() error {
	return e.Err
}

func New(code enum.ErrorCode, language ...string) AppError {
	lang := "en" // default language
	if len(language) > 0 && language[0] == "id" {
		lang = "id"
	}

	var message string
	if lang == "id" {
		message = code.MessageID()
	} else {
		message = code.MessageEN()
	}

	return AppError{
		Code:       code,
		Message:    message,
		StatusCode: code.HTTPStatus(),
		Language:   lang,
	}
}

func NewWithDetails(code enum.ErrorCode, details interface{}, language ...string) AppError {
	err := New(code, language...)
	err.Details = details
	return err
}

func Wrap(err error, code enum.ErrorCode, language ...string) AppError {
	appErr := New(code, language...)
	appErr.Err = err
	return appErr
}

func WrapWithDetails(err error, code enum.ErrorCode, details interface{}, language ...string) AppError {
	appErr := NewWithDetails(code, details, language...)
	appErr.Err = err
	return appErr
}

func IsAppError(err error) (AppError, bool) {
	var appErr AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

type ValidationErrorDetails struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func NewValidationError(details []ValidationErrorDetails, language ...string) AppError {
	return NewWithDetails(ValidationFailed, details, language...)
}

func BadRequest(message string, language ...string) AppError {
	err := New(InvalidRequest, language...)
	if message != "" {
		err.Message = message
	}
	return err
}

func NotFound(message string, language ...string) AppError {
	err := New(ResourceNotFound, language...)
	if message != "" {
		err.Message = message
	}
	return err
}

func UnauthorizedAccess(message string, language ...string) AppError {
	err := New(Unauthorized, language...)
	if message != "" {
		err.Message = message
	}
	return err
}

func InternalError(message string, language ...string) AppError {
	err := New(InternalServerError, language...)
	if message != "" {
		err.Message = message
	}
	return err
}
