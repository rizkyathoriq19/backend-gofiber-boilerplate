package errors

import (
	"net/http"
	"testing"

	"boilerplate-be/internal/pkg/enum"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		code       enum.ErrorCode
		wantStatus int
	}{
		{
			name:       "Unauthorized error",
			code:       enum.Unauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Not found error",
			code:       enum.ResourceNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Validation failed",
			code:       enum.ValidationFailed,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Internal server error",
			code:       enum.InternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Forbidden error",
			code:       enum.Forbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Conflict error",
			code:       enum.Conflict,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code)

			if err.StatusCode != tt.wantStatus {
				t.Errorf("New() StatusCode = %v, want %v", err.StatusCode, tt.wantStatus)
			}

			if err.Code != tt.code {
				t.Errorf("New() Code = %v, want %v", err.Code, tt.code)
			}

			if err.Message == "" {
				t.Error("New() Message should not be empty")
			}
		})
	}
}

func TestNewWithDetails(t *testing.T) {
	details := []ValidationErrorDetails{
		{Field: "email", Message: "invalid email"},
	}

	err := NewWithDetails(enum.ValidationFailed, details)

	if err.Details == nil {
		t.Error("NewWithDetails() Details should not be nil")
	}
}

func TestAppError_Error(t *testing.T) {
	err := New(enum.ResourceNotFound)

	errStr := err.Error()

	if errStr == "" {
		t.Error("AppError.Error() should not return empty string")
	}
}

func TestBadRequest(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "With message",
			message: "custom message",
		},
		{
			name:    "Empty message (uses default)",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BadRequest(tt.message)

			if err.StatusCode != http.StatusBadRequest {
				t.Errorf("BadRequest() StatusCode = %v, want %v", err.StatusCode, http.StatusBadRequest)
			}

			if tt.message != "" && err.Message != tt.message {
				t.Errorf("BadRequest() Message = %v, want %v", err.Message, tt.message)
			}
		})
	}
}

func TestNotFound_Function(t *testing.T) {
	err := NotFound("resource not found")

	if err.StatusCode != http.StatusNotFound {
		t.Errorf("NotFound() StatusCode = %v, want %v", err.StatusCode, http.StatusNotFound)
	}

	if err.Message != "resource not found" {
		t.Errorf("NotFound() Message = %v, want %v", err.Message, "resource not found")
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	err := UnauthorizedAccess("not authenticated")

	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("UnauthorizedAccess() StatusCode = %v, want %v", err.StatusCode, http.StatusUnauthorized)
	}
}

func TestInternalError(t *testing.T) {
	err := InternalError("server error")

	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("InternalError() StatusCode = %v, want %v", err.StatusCode, http.StatusInternalServerError)
	}
}

func TestWrap(t *testing.T) {
	originalErr := &testError{msg: "original error"}

	wrappedErr := Wrap(originalErr, enum.InternalServerError)

	if wrappedErr.Err == nil {
		t.Error("Wrap() should set Err field")
	}

	if wrappedErr.Unwrap() != originalErr {
		t.Error("Wrap() Unwrap should return original error")
	}
}

// testError for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
