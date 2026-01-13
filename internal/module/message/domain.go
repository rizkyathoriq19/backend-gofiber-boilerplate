package message

// MessageRepository defines the interface for message data operations
type MessageRepository interface {
	Create(msg *Message) error
	GetByID(id string) (*Message, error)
	GetByRoomID(roomID string, filter *MessageFilter) ([]*MessageWithDetails, int, error)
	GetByPatientID(patientID string, filter *MessageFilter) ([]*MessageWithDetails, int, error)
	GetByStaffID(staffID string, filter *MessageFilter) ([]*MessageWithDetails, int, error)
	GetUnreadCount(userID string, isStaff bool) (int, error)
	MarkAsRead(id string) error
	MarkAllAsRead(roomID string, userID string) error
	Delete(id string) error
}

// MessageUseCase defines the interface for message business logic
type MessageUseCase interface {
	SendMessage(req *SendMessageRequest, senderID string, isStaff bool) (*Message, error)
	GetMessage(id string) (*Message, error)
	GetRoomMessages(roomID string, filter *MessageFilter) ([]*MessageWithDetails, int, error)
	GetPatientMessages(patientID string, filter *MessageFilter) ([]*MessageWithDetails, int, error)
	GetStaffMessages(staffID string, filter *MessageFilter) ([]*MessageWithDetails, int, error)
	GetUnreadCount(userID string, isStaff bool) (int, error)
	MarkAsRead(id string) error
	MarkAllAsRead(roomID string, userID string) error
	DeleteMessage(id string) error
}
