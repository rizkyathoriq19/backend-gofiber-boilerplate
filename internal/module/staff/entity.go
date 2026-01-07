package staff

import (
	"time"
)

// StaffType represents the type of staff member
type StaffType string

const (
	StaffTypeNurse   StaffType = "nurse"
	StaffTypeDoctor  StaffType = "doctor"
	StaffTypeManager StaffType = "manager"
	StaffTypeAdmin   StaffType = "admin"
)

// ShiftType represents the shift type
type ShiftType string

const (
	ShiftMorning   ShiftType = "morning"
	ShiftAfternoon ShiftType = "afternoon"
	ShiftNight     ShiftType = "night"
	ShiftOnCall    ShiftType = "on_call"
)

// Staff represents a staff member in the hospital
type Staff struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	EmployeeID string    `json:"employee_id" db:"employee_id"`
	Type       StaffType `json:"type" db:"type"`
	Department string    `json:"department" db:"department"`
	Shift      ShiftType `json:"shift" db:"shift"`
	OnDuty     bool      `json:"on_duty" db:"on_duty"`
	Phone      string    `json:"phone" db:"phone"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// StaffWithUser includes user details
type StaffWithUser struct {
	Staff
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// StaffRoomAssignment represents a staff-room assignment
type StaffRoomAssignment struct {
	ID         string    `json:"id" db:"id"`
	StaffID    string    `json:"staff_id" db:"staff_id"`
	RoomID     string    `json:"room_id" db:"room_id"`
	IsPrimary  bool      `json:"is_primary" db:"is_primary"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

// StaffFilter represents filters for querying staff
type StaffFilter struct {
	Type       StaffType `query:"type"`
	Department string    `query:"department"`
	Shift      ShiftType `query:"shift"`
	OnDuty     *bool     `query:"on_duty"`
	Page       int       `query:"page"`
	Limit      int       `query:"limit"`
}
