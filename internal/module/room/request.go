package room

// CreateRoomRequest represents the request to create a room
type CreateRoomRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=100"`
	Type     RoomType `json:"type" validate:"required,oneof=patient_room nurse_station icu emergency operating_room"`
	Floor    string   `json:"floor" validate:"required,max=20"`
	Building string   `json:"building" validate:"max=100"`
	Capacity int      `json:"capacity" validate:"min=0,max=50"`
}

// UpdateRoomRequest represents the request to update a room
type UpdateRoomRequest struct {
	Name     string   `json:"name" validate:"omitempty,min=2,max=100"`
	Type     RoomType `json:"type" validate:"omitempty,oneof=patient_room nurse_station icu emergency operating_room"`
	Floor    string   `json:"floor" validate:"omitempty,max=20"`
	Building string   `json:"building" validate:"omitempty,max=100"`
	Capacity int      `json:"capacity" validate:"omitempty,min=0,max=50"`
	IsActive *bool    `json:"is_active"`
}
