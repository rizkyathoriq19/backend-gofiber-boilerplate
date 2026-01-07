package staff

// StaffRepository defines the interface for staff data operations
type StaffRepository interface {
	Create(staff *Staff) error
	GetByID(id string) (*Staff, error)
	GetByUserID(userID string) (*Staff, error)
	GetByEmployeeID(employeeID string) (*Staff, error)
	GetAll(filter *StaffFilter) ([]*StaffWithUser, int, error)
	GetOnDutyByType(staffType StaffType) ([]*StaffWithUser, error)
	GetOnDutyByRoom(roomID string) ([]*StaffWithUser, error)
	Update(staff *Staff) error
	UpdateOnDutyStatus(id string, onDuty bool) error
	Delete(id string) error

	// Room assignments
	AssignToRoom(staffID, roomID string, isPrimary bool) error
	RemoveFromRoom(staffID, roomID string) error
	GetRoomAssignments(staffID string) ([]*StaffRoomAssignment, error)
	GetStaffByRoom(roomID string) ([]*StaffWithUser, error)
	IsAssignedToRoom(staffID, roomID string) (bool, error)
}

// StaffUseCase defines the interface for staff business logic
type StaffUseCase interface {
	CreateStaff(req *CreateStaffRequest) (*Staff, error)
	GetStaff(id string) (*Staff, error)
	GetStaffByUserID(userID string) (*Staff, error)
	GetAllStaff(filter *StaffFilter) ([]*StaffWithUser, int, error)
	GetOnDutyStaff(staffType StaffType) ([]*StaffWithUser, error)
	GetOnDutyStaffByRoom(roomID string) ([]*StaffWithUser, error)
	UpdateStaff(id string, req *UpdateStaffRequest) (*Staff, error)
	UpdateShift(id string, shift ShiftType) error
	ToggleOnDuty(id string) error
	DeleteStaff(id string) error

	// Room assignments
	AssignToRoom(staffID, roomID string, isPrimary bool) error
	RemoveFromRoom(staffID, roomID string) error
	GetRoomAssignments(staffID string) ([]*StaffRoomAssignment, error)
	IsAssignedToRoom(staffID, roomID string) (bool, error)
}
