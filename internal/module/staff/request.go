package staff

// CreateStaffRequest represents the request to create a staff member
type CreateStaffRequest struct {
	UserID     string    `json:"user_id" validate:"required,uuid"`
	EmployeeID string    `json:"employee_id" validate:"required,min=1,max=50"`
	Type       StaffType `json:"type" validate:"required,oneof=nurse doctor manager admin"`
	Department string    `json:"department" validate:"omitempty,max=100"`
	Shift      ShiftType `json:"shift" validate:"omitempty,oneof=morning afternoon night on_call"`
	Phone      string    `json:"phone" validate:"omitempty,max=20"`
}

// UpdateStaffRequest represents the request to update a staff member
type UpdateStaffRequest struct {
	Department string    `json:"department" validate:"omitempty,max=100"`
	Shift      ShiftType `json:"shift" validate:"omitempty,oneof=morning afternoon night on_call"`
	Phone      string    `json:"phone" validate:"omitempty,max=20"`
}

// UpdateShiftRequest represents the request to update staff shift
type UpdateShiftRequest struct {
	Shift ShiftType `json:"shift" validate:"required,oneof=morning afternoon night on_call"`
}

// AssignRoomRequest represents the request to assign staff to a room
type AssignRoomRequest struct {
	RoomID    string `json:"room_id" validate:"required,uuid"`
	IsPrimary bool   `json:"is_primary"`
}
