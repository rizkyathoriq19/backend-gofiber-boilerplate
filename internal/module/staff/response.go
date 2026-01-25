package staff

import (
	"time"
)

// StaffResponse represents the staff response
type StaffResponse struct {
	ID         string    `json:"uuid"`
	UserID     string    `json:"user_uuid"`
	EmployeeID string    `json:"employee_id"`
	Type       StaffType `json:"type"`
	Department string    `json:"department"`
	Shift      ShiftType `json:"shift"`
	OnDuty     bool      `json:"on_duty"`
	Phone      string    `json:"phone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StaffWithUserResponse includes user details
type StaffWithUserResponse struct {
	StaffResponse
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// StaffListResponse represents paginated staff list
type StaffListResponse struct {
	Staff      []*StaffWithUserResponse `json:"staff"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalPages int                      `json:"total_pages"`
}

// RoomAssignmentResponse represents a room assignment
type RoomAssignmentResponse struct {
	ID         string    `json:"uuid"`
	RoomID     string    `json:"room_uuid"`
	IsPrimary  bool      `json:"is_primary"`
	AssignedAt time.Time `json:"assigned_at"`
}

// ToResponse converts Staff entity to StaffResponse
func (s *Staff) ToResponse() *StaffResponse {
	return &StaffResponse{
		ID:         s.ID,
		UserID:     s.UserID,
		EmployeeID: s.EmployeeID,
		Type:       s.Type,
		Department: s.Department,
		Shift:      s.Shift,
		OnDuty:     s.OnDuty,
		Phone:      s.Phone,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

// ToResponse converts StaffWithUser to StaffWithUserResponse
func (s *StaffWithUser) ToResponse() *StaffWithUserResponse {
	return &StaffWithUserResponse{
		StaffResponse: *s.Staff.ToResponse(),
		UserName:      s.UserName,
		UserEmail:     s.UserEmail,
	}
}
