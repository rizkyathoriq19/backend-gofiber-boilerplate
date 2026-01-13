package staff

import (
	"boilerplate-be/internal/shared/errors"
)

type staffUseCase struct {
	staffRepo StaffRepository
}

// NewStaffUseCase creates a new staff use case
func NewStaffUseCase(staffRepo StaffRepository) StaffUseCase {
	return &staffUseCase{
		staffRepo: staffRepo,
	}
}

// CreateStaff creates a new staff member
func (u *staffUseCase) CreateStaff(req *CreateStaffRequest) (*Staff, error) {
	// Check if employee ID already exists
	existing, err := u.staffRepo.GetByEmployeeID(req.EmployeeID)
	if err == nil && existing != nil {
		return nil, errors.New(errors.ResourceAlreadyExists)
	}

	// Check if user already has a staff profile
	existingByUser, err := u.staffRepo.GetByUserID(req.UserID)
	if err == nil && existingByUser != nil {
		return nil, errors.New(errors.ResourceAlreadyExists)
	}

	staff := &Staff{
		UserID:     req.UserID,
		EmployeeID: req.EmployeeID,
		Type:       req.Type,
		Department: req.Department,
		Shift:      req.Shift,
		Phone:      req.Phone,
	}

	if staff.Shift == "" {
		staff.Shift = ShiftMorning
	}

	if err := u.staffRepo.Create(staff); err != nil {
		return nil, err
	}

	return staff, nil
}

// GetStaff gets a staff member by ID
func (u *staffUseCase) GetStaff(id string) (*Staff, error) {
	return u.staffRepo.GetByID(id)
}

// GetStaffByUserID gets a staff member by user ID
func (u *staffUseCase) GetStaffByUserID(userID string) (*Staff, error) {
	return u.staffRepo.GetByUserID(userID)
}

// GetAllStaff gets all staff with filters
func (u *staffUseCase) GetAllStaff(filter *StaffFilter) ([]*StaffWithUser, int, error) {
	return u.staffRepo.GetAll(filter)
}

// GetOnDutyStaff gets on-duty staff by type
func (u *staffUseCase) GetOnDutyStaff(staffType StaffType) ([]*StaffWithUser, error) {
	return u.staffRepo.GetOnDutyByType(staffType)
}

// GetOnDutyStaffByRoom gets on-duty staff assigned to a room
func (u *staffUseCase) GetOnDutyStaffByRoom(roomID string) ([]*StaffWithUser, error) {
	return u.staffRepo.GetOnDutyByRoom(roomID)
}

// UpdateStaff updates a staff member
func (u *staffUseCase) UpdateStaff(id string, req *UpdateStaffRequest) (*Staff, error) {
	staff, err := u.staffRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Department != "" {
		staff.Department = req.Department
	}
	if req.Shift != "" {
		staff.Shift = req.Shift
	}
	if req.Phone != "" {
		staff.Phone = req.Phone
	}

	if err := u.staffRepo.Update(staff); err != nil {
		return nil, err
	}

	return staff, nil
}

// UpdateShift updates a staff member's shift
func (u *staffUseCase) UpdateShift(id string, shift ShiftType) error {
	staff, err := u.staffRepo.GetByID(id)
	if err != nil {
		return err
	}

	staff.Shift = shift
	return u.staffRepo.Update(staff)
}

// ToggleOnDuty toggles the on-duty status
func (u *staffUseCase) ToggleOnDuty(id string) error {
	staff, err := u.staffRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.staffRepo.UpdateOnDutyStatus(id, !staff.OnDuty)
}

// DeleteStaff deletes a staff member
func (u *staffUseCase) DeleteStaff(id string) error {
	_, err := u.staffRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.staffRepo.Delete(id)
}

// AssignToRoom assigns a staff member to a room
func (u *staffUseCase) AssignToRoom(staffID, roomID string, isPrimary bool) error {
	_, err := u.staffRepo.GetByID(staffID)
	if err != nil {
		return err
	}

	return u.staffRepo.AssignToRoom(staffID, roomID, isPrimary)
}

// RemoveFromRoom removes a staff member from a room
func (u *staffUseCase) RemoveFromRoom(staffID, roomID string) error {
	return u.staffRepo.RemoveFromRoom(staffID, roomID)
}

// GetRoomAssignments gets all room assignments for a staff member
func (u *staffUseCase) GetRoomAssignments(staffID string) ([]*StaffRoomAssignment, error) {
	return u.staffRepo.GetRoomAssignments(staffID)
}

// IsAssignedToRoom checks if a staff member is assigned to a room
func (u *staffUseCase) IsAssignedToRoom(staffID, roomID string) (bool, error) {
	return u.staffRepo.IsAssignedToRoom(staffID, roomID)
}
