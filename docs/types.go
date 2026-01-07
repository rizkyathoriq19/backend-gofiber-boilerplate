// Package docs contains Swagger definitions for the API.
package docs

import "time"

// SuccessResponse represents a successful API response
// @Description Standard success response wrapper
type SuccessResponse struct {
	Status    bool        `json:"status" example:"true"`
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"Operation successful"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse represents an error API response
// @Description Standard error response wrapper
type ErrorResponse struct {
	Status    bool      `json:"status" example:"false"`
	Code      int       `json:"code" example:"400"`
	Message   string    `json:"message" example:"Error occurred"`
	ErrorCode int       `json:"error_code,omitempty" example:"-1001"`
	Timestamp time.Time `json:"timestamp"`
}

// UserResponse represents user data in responses
// @Description User information
type UserResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john@example.com"`
	Role      string    `json:"role" example:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthResponse represents authentication response with user and tokens
// @Description Authentication response with user data and tokens
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string       `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string       `json:"token_type" example:"Bearer"`
	ExpiresIn    int64        `json:"expires_in" example:"86400"`
}

// TokenResponse represents token refresh response
// @Description Token response with access and refresh tokens
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in" example:"86400"`
}

// RoleResponse represents role data
// @Description Role information
type RoleResponse struct {
	ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"admin"`
	Description string    `json:"description,omitempty" example:"Administrator role"`
	CreatedAt   time.Time `json:"created_at"`
}

// PermissionResponse represents permission data
// @Description Permission information
type PermissionResponse struct {
	ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"users:read"`
	Description string    `json:"description,omitempty" example:"View users"`
	Resource    string    `json:"resource" example:"users"`
	Action      string    `json:"action" example:"read"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRolesResponse represents user with their roles
// @Description User roles response
type UserRolesResponse struct {
	UserID string         `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Roles  []RoleResponse `json:"roles"`
}

// RegisterRequest represents registration payload
// @Description User registration request
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=6"`
	Name     string `json:"name" example:"John Doe" validate:"required,min=2"`
}

// LoginRequest represents login payload
// @Description User login request
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required"`
}

// RefreshTokenRequest represents refresh token payload
// @Description Refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..." validate:"required"`
}

// UpdateProfileRequest represents profile update payload
// @Description Profile update request
type UpdateProfileRequest struct {
	Name string `json:"name" example:"John Updated" validate:"omitempty,min=2"`
}

// CreateRoleRequest represents role creation payload
// @Description Role creation request
type CreateRoleRequest struct {
	Name        string `json:"name" example:"moderator" validate:"required,min=2,max=50"`
	Description string `json:"description" example:"Content moderator" validate:"max=255"`
}

// AssignRoleRequest represents role assignment payload
// @Description Role assignment request
type AssignRoleRequest struct {
	RoleID string `json:"role_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required,uuid"`
}

// AssignPermissionRequest represents permission assignment payload
// @Description Permission assignment request
type AssignPermissionRequest struct {
	PermissionID string `json:"permission_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required,uuid"`
}
