package room

// RoomRepository defines the interface for room data operations
type RoomRepository interface {
	Create(room *Room) error
	GetByID(id string) (*Room, error)
	GetAll(filter *RoomFilter) ([]*Room, int, error)
	Update(room *Room) error
	Delete(id string) error
}

// RoomUseCase defines the interface for room business logic
type RoomUseCase interface {
	CreateRoom(req *CreateRoomRequest) (*Room, error)
	GetRoom(id string) (*Room, error)
	GetRooms(filter *RoomFilter) ([]*Room, int, error)
	UpdateRoom(id string, req *UpdateRoomRequest) (*Room, error)
	DeleteRoom(id string) error
}
