package room

import (
	"boilerplate-be/internal/shared/errors"
)

type roomUseCase struct {
	roomRepo RoomRepository
}

// NewRoomUseCase creates a new room use case
func NewRoomUseCase(roomRepo RoomRepository) RoomUseCase {
	return &roomUseCase{
		roomRepo: roomRepo,
	}
}

// CreateRoom creates a new room
func (u *roomUseCase) CreateRoom(req *CreateRoomRequest) (*Room, error) {
	room := &Room{
		Name:     req.Name,
		Type:     req.Type,
		Floor:    req.Floor,
		Building: req.Building,
		Capacity: req.Capacity,
	}

	if room.Capacity == 0 {
		room.Capacity = 1
	}

	if err := u.roomRepo.Create(room); err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom gets a room by ID
func (u *roomUseCase) GetRoom(id string) (*Room, error) {
	room, err := u.roomRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return room, nil
}

// GetRooms gets all rooms with filters
func (u *roomUseCase) GetRooms(filter *RoomFilter) ([]*Room, int, error) {
	rooms, total, err := u.roomRepo.GetAll(filter)
	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// UpdateRoom updates a room
func (u *roomUseCase) UpdateRoom(id string, req *UpdateRoomRequest) (*Room, error) {
	// Get existing room
	room, err := u.roomRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != "" {
		room.Name = req.Name
	}
	if req.Type != "" {
		room.Type = req.Type
	}
	if req.Floor != "" {
		room.Floor = req.Floor
	}
	if req.Building != "" {
		room.Building = req.Building
	}
	if req.Capacity > 0 {
		room.Capacity = req.Capacity
	}
	if req.IsActive != nil {
		room.IsActive = *req.IsActive
	}

	if err := u.roomRepo.Update(room); err != nil {
		return nil, err
	}

	return room, nil
}

// DeleteRoom deletes a room
func (u *roomUseCase) DeleteRoom(id string) error {
	// Check if room exists
	_, err := u.roomRepo.GetByID(id)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			if appErr.Code == errors.ResourceNotFound {
				return errors.New(errors.ResourceNotFound)
			}
		}
		return err
	}

	return u.roomRepo.Delete(id)
}
