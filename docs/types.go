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
	Email     string    `json:"email" example:"patient@example.com"`
	Role      string    `json:"role" example:"patient"`
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
	Name        string    `json:"name" example:"superadmin"`
	Description string    `json:"description,omitempty" example:"Full system access"`
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
	Email    string `json:"email" example:"patient@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required,min=6"`
	Name     string `json:"name" example:"John Doe" validate:"required,min=2"`
}

// LoginRequest represents login payload
// @Description User login request
type LoginRequest struct {
	Email    string `json:"email" example:"patient@example.com" validate:"required,email"`
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

// ========== MEDIPROMPT Types ==========

// RoomDocResponse represents room data
// @Description Room information
type RoomDocResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name" example:"Room 101"`
	Type      string    `json:"type" example:"patient_room"`
	Floor     string    `json:"floor" example:"1"`
	Building  string    `json:"building" example:"Main"`
	Capacity  int       `json:"capacity" example:"2"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RoomListDocResponse represents paginated room list
// @Description Paginated room list
type RoomListDocResponse struct {
	Rooms      []RoomDocResponse `json:"rooms"`
	Total      int               `json:"total" example:"100"`
	Page       int               `json:"page" example:"1"`
	Limit      int               `json:"limit" example:"10"`
	TotalPages int               `json:"total_pages" example:"10"`
}

// DeviceDocResponse represents device data
// @Description Device information
type DeviceDocResponse struct {
	ID            string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoomID        *string                `json:"room_id"`
	Type          string                 `json:"type" example:"microphone"`
	SerialNumber  string                 `json:"serial_number" example:"MIC-001"`
	Name          string                 `json:"name" example:"Room 101 Microphone"`
	Status        string                 `json:"status" example:"online"`
	Config        map[string]interface{} `json:"config"`
	LastHeartbeat *time.Time             `json:"last_heartbeat"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// DeviceWithAPIKeyDocResponse includes API key
// @Description Device with API key (only on registration)
type DeviceWithAPIKeyDocResponse struct {
	DeviceDocResponse
	APIKey string `json:"api_key" example:"mp_abc123def456..."`
}

// DeviceListDocResponse represents paginated device list
// @Description Paginated device list
type DeviceListDocResponse struct {
	Devices    []DeviceDocResponse `json:"devices"`
	Total      int                 `json:"total" example:"50"`
	Page       int                 `json:"page" example:"1"`
	Limit      int                 `json:"limit" example:"10"`
	TotalPages int                 `json:"total_pages" example:"5"`
}

// StaffDocResponse represents staff data
// @Description Staff information
type StaffDocResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID     string    `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	EmployeeID string    `json:"employee_id" example:"EMP-001"`
	Type       string    `json:"type" example:"nurse"`
	Department string    `json:"department" example:"ICU"`
	Shift      string    `json:"shift" example:"morning"`
	OnDuty     bool      `json:"on_duty" example:"true"`
	Phone      string    `json:"phone" example:"+62812345678"`
	UserName   string    `json:"user_name" example:"John Doe"`
	UserEmail  string    `json:"user_email" example:"john@hospital.com"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StaffListDocResponse represents paginated staff list
// @Description Paginated staff list
type StaffListDocResponse struct {
	Staff      []StaffDocResponse `json:"staff"`
	Total      int                `json:"total" example:"25"`
	Page       int                `json:"page" example:"1"`
	Limit      int                `json:"limit" example:"10"`
	TotalPages int                `json:"total_pages" example:"3"`
}

// RoomAssignmentDocResponse represents room assignment
// @Description Staff room assignment
type RoomAssignmentDocResponse struct {
	ID         string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoomID     string    `json:"room_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	IsPrimary  bool      `json:"is_primary" example:"true"`
	AssignedAt time.Time `json:"assigned_at"`
}

// PatientDocResponse represents patient data
// @Description Patient information
type PatientDocResponse struct {
	ID                  string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoomID              *string    `json:"room_id"`
	MedicalRecordNumber string     `json:"medical_record_number" example:"MRN-001"`
	Name                string     `json:"name" example:"Jane Doe"`
	Gender              string     `json:"gender" example:"female"`
	ConditionLevel      string     `json:"condition_level" example:"stable"`
	Diagnosis           string     `json:"diagnosis" example:"Pneumonia"`
	RoomName            string     `json:"room_name,omitempty" example:"Room 101"`
	RoomType            string     `json:"room_type,omitempty" example:"patient_room"`
	AdmissionDate       time.Time  `json:"admission_date"`
	DischargeDate       *time.Time `json:"discharge_date"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PatientListDocResponse represents paginated patient list
// @Description Paginated patient list
type PatientListDocResponse struct {
	Patients   []PatientDocResponse `json:"patients"`
	Total      int                  `json:"total" example:"200"`
	Page       int                  `json:"page" example:"1"`
	Limit      int                  `json:"limit" example:"10"`
	TotalPages int                  `json:"total_pages" example:"20"`
}

// AlertDocResponse represents alert data
// @Description Alert information
type AlertDocResponse struct {
	ID                string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoomID            string     `json:"room_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PatientID         *string    `json:"patient_id"`
	DeviceID          *string    `json:"device_id"`
	AssignedStaffID   *string    `json:"assigned_staff_id"`
	Type              string     `json:"type" example:"voice_call"`
	Priority          string     `json:"priority" example:"high"`
	Status            string     `json:"status" example:"pending"`
	Message           string     `json:"message" example:"Patient requesting assistance"`
	DetectedKeywords  []string   `json:"detected_keywords,omitempty"`
	EscalationCount   int        `json:"escalation_count" example:"0"`
	RoomName          string     `json:"room_name" example:"Room 101"`
	PatientName       string     `json:"patient_name,omitempty" example:"Jane Doe"`
	AssignedStaff     string     `json:"assigned_staff,omitempty" example:"John Nurse"`
	CreatedAt         time.Time  `json:"created_at"`
	AcknowledgedAt    *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// AlertListDocResponse represents paginated alert list
// @Description Paginated alert list
type AlertListDocResponse struct {
	Alerts     []AlertDocResponse `json:"alerts"`
	Total      int                `json:"total" example:"50"`
	Page       int                `json:"page" example:"1"`
	Limit      int                `json:"limit" example:"10"`
	TotalPages int                `json:"total_pages" example:"5"`
}

// AlertHistoryDocResponse represents alert history entry
// @Description Alert history entry
type AlertHistoryDocResponse struct {
	ID             string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Action         string    `json:"action" example:"acknowledged"`
	PreviousStatus string    `json:"previous_status" example:"pending"`
	NewStatus      string    `json:"new_status" example:"acknowledged"`
	Notes          string    `json:"notes,omitempty" example:"Nurse on the way"`
	CreatedAt      time.Time `json:"created_at"`
}

