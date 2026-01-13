package enum

import "net/http"

func (ec ErrorCode) HTTPStatus() int {
	switch ec {
	// Success cases
	case Success:
		return http.StatusOK

	// Client Errors (400-499)
	case InvalidRequest, InvalidRequestBody, MissingRequiredField, 
		InvalidFormat, ValidationFailed:
		return http.StatusBadRequest

	case InvalidCredentials, Unauthorized, InvalidToken, TokenExpired:
		return http.StatusUnauthorized

	case Forbidden:
		return http.StatusForbidden

	case ResourceNotFound, NoDataFound, DataNotFound, AccountNotFound:
		return http.StatusNotFound

	case Conflict, UsernameExists, EmailExists:
		return http.StatusConflict

	case InvalidUsername, InvalidEmail, PasswordMismatch, AccountInactive:
		return http.StatusUnprocessableEntity

	case RateLimitExceeded:
		return http.StatusTooManyRequests

	// Server Errors (500-599)
	case InternalServerError, DatabaseError, CacheError, ConfigurationError,
		DatabaseConnectionFailed, DatabaseQueryFailed, DatabaseInsertFailed,
		DatabaseUpdateFailed, DatabaseDeleteFailed, DatabaseScanFailed,
		CacheConnectionFailed, CacheStoreFailed, CacheRetrieveFailed,
		CacheDeleteFailed, TokenGenerationFailed, PasswordHashFailed:
		return http.StatusInternalServerError

	case ExternalServiceError, ServiceUnavailable, AuthServiceUnavailable:
		return http.StatusServiceUnavailable

	default:
		return http.StatusInternalServerError
	}
}

func (ec ErrorCode) HTTPStatusText() string {
	return http.StatusText(ec.HTTPStatus())
}

// IsClientError returns true if the error code represents a client error (4xx)
func (ec ErrorCode) IsClientError() bool {
	status := ec.HTTPStatus()
	return status >= 400 && status < 500
}

// IsServerError returns true if the error code represents a server error (5xx)
func (ec ErrorCode) IsServerError() bool {
	status := ec.HTTPStatus()
	return status >= 500 && status < 600
}

// IsSuccess returns true if the error code represents success (2xx)
func (ec ErrorCode) IsSuccess() bool {
	status := ec.HTTPStatus()
	return status >= 200 && status < 300
}
