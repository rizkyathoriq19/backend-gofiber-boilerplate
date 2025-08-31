package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode merepresentasikan kode error yang sudah ditentukan
type ErrorCode int

const (
	// General (0-99)
	NotSetYet ErrorCode = iota
	SuccessOnRequest
	NoDataFound
	CantFetchCreatedData

	// Server Errors (10000-10999)
	ServerError
	ServerErrorCantGeneratePassword
	ServerErrorFailedGenerateToken
	ServerErrorRedis
	ServerErrorRedisCantStore
	ServerCantInsertUserData
	ServerCantScanUserData

	// Authentication & Authorization (20000-20999)
	InvalidCredential
	Unauthorized
	DontHavePermissionToAccess
	ValidatedCredentialsSuccess

	// User & Email Issues (30000-30999)
	UsernameHasBeenUsed
	EmailHasBeenUsed
	CantPickUsername
	EmailContainInvalidCharacter

	// Request & Parsing Errors (40000-40999)
	BodyRequestError
	CantParseRequestBody
	ImportantBodyParserNotIncluded
	UnableGetProfile
	NoIdDataSearch

	// Validation Errors (50000-50999)
	ValidationError

	// JWT Errors (60000-60999)
	InvalidToken

	// Rate Limit Errors (70000-70999)
	RateLimitExceeded
)

// Value mengembalikan nilai numerik dari error code
func (ec ErrorCode) Value() int {
	return [...]int{
		// General
		0,    // NotSetYet
		-100, // SuccessOnRequest
		-101, // NoDataFound
		-102, // CantFetchCreatedData

		// Server Errors
		-10000, // ServerError
		-10010, // ServerErrorCantGeneratePassword
		-10020, // ServerErrorFailedGenerateToken
		-10030, // ServerErrorRedis
		-10031, // ServerErrorRedisCantStore
		-10032, // ServerCantInsertUserData
		-10033, // ServerCantScanUserData

		// Authentication & Authorization
		-20000, // InvalidCredential
		-20010, // Unauthorized
		-20020, // DontHavePermissionToAccess
		-20030, // ValidatedCredentialsSuccess

		// User & Email Issues
		-30000, // UsernameHasBeenUsed
		-30010, // EmailHasBeenUsed
		-30020, // CantPickUsername
		-30030, // EmailContainInvalidCharacter

		// Request & Parsing Errors
		-40000, // BodyRequestError
		-40010, // CantParseRequestBody
		-40020, // ImportantBodyParserNotIncluded
		-40030, // UnableGetProfile
		-40031, // NoIdDataSearch

		// Validation Errors
		-50000, // ValidationError

		// JWT Errors
		-60000, // InvalidToken

		// Rate Limit Errors
		-70000, // RateLimitExceeded
	}[ec]
}

// MessageID mengembalikan pesan error dalam Bahasa Indonesia
func (ec ErrorCode) MessageID() string {
	return [...]string{
		// General
		"Kode error belum ditentukan",           // NotSetYet
		"Permintaan berhasil diproses",          // SuccessOnRequest
		"Data tidak ditemukan",                  // NoDataFound
		"Gagal mengambil data yang baru dibuat", // CantFetchCreatedData

		// Server Errors
		"Terjadi kesalahan pada server, coba lagi nanti", // ServerError
		"Gagal membuat kata sandi, coba lagi nanti",      // ServerErrorCantGeneratePassword
		"Gagal membuat token, coba lagi nanti",           // ServerErrorFailedGenerateToken
		"Kesalahan pada server Redis, coba lagi nanti",   // ServerErrorRedis
		"Gagal menyimpan data di Redis",                  // ServerErrorRedisCantStore
		"Gagal menyimpan data pengguna",                  // ServerCantInsertUserData
		"Gagal melakukan analisis data pengguna",         // ServerCantScanUserData

		// Authentication & Authorization
		"Kredensial tidak valid",              // InvalidCredential
		"Permintaan tidak sah",                // Unauthorized
		"Tidak memiliki izin untuk mengakses", // DontHavePermissionToAccess
		"Kredensial berhasil divalidasi",      // ValidatedCredentialsSuccess

		// User & Email Issues
		"Nama pengguna sudah digunakan",                  // UsernameHasBeenUsed
		"Email sudah digunakan",                          // EmailHasBeenUsed
		"Tidak dapat menggunakan nama pengguna tersebut", // CantPickUsername
		"Email mengandung karakter tidak valid",          // EmailContainInvalidCharacter

		// Request & Parsing Errors
		"Format permintaan tidak sesuai", // BodyRequestError
		"Gagal memproses isi permintaan", // CantParseRequestBody
		"Data penting tidak disertakan",  // ImportantBodyParserNotIncluded
		"Gagal mengambil data profil",    // UnableGetProfile
		"Tidak ada data id pencarian",    // NoIdDataSearch

		// Validation Errors
		"Validasi gagal", // ValidationError

		// JWT Errors
		"Kredensial tidak valid", // InvalidToken

		// Rate Limit Errors
		"Rate limit telah tercapai", // RateLimitExceeded
	}[ec]
}

// MessageEN mengembalikan pesan error dalam Bahasa Inggris
func (ec ErrorCode) MessageEN() string {
	return [...]string{
		// General
		"Error code not set yet",         // NotSetYet
		"Request completed successfully", // SuccessOnRequest
		"No data found",                  // NoDataFound
		"Failed to fetch created data",   // CantFetchCreatedData

		// Server Errors
		"Server error, please try again later",         // ServerError
		"Failed to generate password, try again later", // ServerErrorCantGeneratePassword
		"Failed to generate token, try again later",    // ServerErrorFailedGenerateToken
		"Redis server error, try again later",          // ServerErrorRedis
		"Failed to store data in Redis",                // ServerErrorRedisCantStore
		"Failed store user data",                       // ServerCantInsertUserData
		"Failed to scan user data",                     // ServerCantScanUserData

		// Authentication & Authorization
		"Invalid credentials provided",         // InvalidCredential
		"Unauthorized request",                 // Unauthorized
		"Permission denied to access resource", // DontHavePermissionToAccess
		"Credentials successfully validated",   // ValidatedCredentialsSuccess

		// User & Email Issues
		"Username has already been used",    // UsernameHasBeenUsed
		"Email has already been used",       // EmailHasBeenUsed
		"Cannot use the selected username",  // CantPickUsername
		"Email contains invalid characters", // EmailContainInvalidCharacter

		// Request & Parsing Errors
		"Invalid request format",          // BodyRequestError
		"Failed to parse request body",    // CantParseRequestBody
		"Important data not included",     // ImportantBodyParserNotIncluded
		"Unable to retrieve profile data", // UnableGetProfile
		"No id search",                    // NoIdDataSearch

		// Validation Errors
		"Validation failed", // ValidationError

		// JWT Errors
		"Invalid credentials", // InvalidToken

		// Rate Limit Errors
		"Rate limit exceeded", // RateLimitExceeded
	}[ec]
}

// HTTPStatus mengembalikan status HTTP yang sesuai untuk error code
func (ec ErrorCode) HTTPStatus() int {
	switch {
	case ec >= ServerError && ec <= ServerCantScanUserData:
		return http.StatusInternalServerError
	case ec >= InvalidCredential && ec <= DontHavePermissionToAccess:
		return http.StatusUnauthorized
	case ec >= UsernameHasBeenUsed && ec <= EmailContainInvalidCharacter:
		return http.StatusBadRequest
	case ec >= BodyRequestError && ec <= NoIdDataSearch:
		return http.StatusBadRequest
	case ec == ValidationError:
		return http.StatusBadRequest
	case ec == InvalidToken:
		return http.StatusUnauthorized
	case ec == RateLimitExceeded:
		return http.StatusTooManyRequests
	case ec == NoDataFound:
		return http.StatusNotFound
	case ec == SuccessOnRequest:
		return http.StatusOK
	default:
		return http.StatusInternalServerError
	}
}

// AppError merepresentasikan error aplikasi
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	StatusCode int       `json:"-"`
	Details    any       `json:"details,omitempty"`
	Err        error     `json:"-"`
	Language   string    `json:"-"`
}

// Error mengimplementasikan interface error
func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap mengembalikan error yang dibungkus
func (e AppError) Unwrap() error {
	return e.Err
}

// New membuat AppError baru berdasarkan ErrorCode
func New(code ErrorCode, language ...string) AppError {
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
	}
}

// NewWithDetails membuat AppError baru dengan details tambahan
func NewWithDetails(code ErrorCode, details any, language ...string) AppError {
	err := New(code, language...)
	err.Details = details
	return err
}

// Wrap membungkus error lain dengan AppError
func Wrap(err error, code ErrorCode, language ...string) AppError {
	appErr := New(code, language...)
	appErr.Err = err
	return appErr
}

// WrapWithDetails membungkus error lain dengan AppError dan details
func WrapWithDetails(err error, code ErrorCode, details any, language ...string) AppError {
	appErr := NewWithDetails(code, details, language...)
	appErr.Err = err
	return appErr
}

// IsAppError memeriksa apakah error adalah AppError
func IsAppError(err error) (AppError, bool) {
	var appErr AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

// ValidationErrorDetails merepresentasikan detail error validasi
type ValidationErrorDetails struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationError membuat error validasi dengan details
func NewValidationError(details []ValidationErrorDetails, language ...string) AppError {
	return NewWithDetails(ValidationError, details, language...)
}

// Predefined errors untuk kemudahan penggunaan
var (
	// General
	ErrNotSetYet         = New(NotSetYet)
	ErrSuccessOnRequest  = New(SuccessOnRequest)
	ErrNoDataFound       = New(NoDataFound)
	ErrCantFetchCreatedData = New(CantFetchCreatedData)

	// Server Errors
	ErrServerError               = New(ServerError)
	ErrCantGeneratePassword      = New(ServerErrorCantGeneratePassword)
	ErrFailedGenerateToken       = New(ServerErrorFailedGenerateToken)
	ErrRedis                     = New(ServerErrorRedis)
	ErrRedisCantStore            = New(ServerErrorRedisCantStore)
	ErrCantInsertUserData        = New(ServerCantInsertUserData)
	ErrCantScanUserData          = New(ServerCantScanUserData)

	// Authentication & Authorization
	ErrInvalidCredential         = New(InvalidCredential)
	ErrUnauthorized              = New(Unauthorized)
	ErrDontHavePermissionToAccess = New(DontHavePermissionToAccess)
	ErrValidatedCredentialsSuccess = New(ValidatedCredentialsSuccess)

	// User & Email Issues	
	ErrUsernameHasBeenUsed       = New(UsernameHasBeenUsed)
	ErrEmailHasBeenUsed          = New(EmailHasBeenUsed)
	ErrCantPickUsername          = New(CantPickUsername)
	ErrEmailContainInvalidCharacter = New(EmailContainInvalidCharacter)

	// Request & Parsing Errors
	ErrBodyRequestError          = New(BodyRequestError)
	ErrCantParseRequestBody      = New(CantParseRequestBody)
	ErrImportantBodyParserNotIncluded = New(ImportantBodyParserNotIncluded)
	ErrUnableGetProfile          = New(UnableGetProfile)
	ErrNoIdDataSearch            = New(NoIdDataSearch)

	// Validation Errors
	ErrValidation                = New(ValidationError)

	ErrInvalidToken              = New(InvalidToken)

	ErrRateLimitExceeded         = New(RateLimitExceeded)
)