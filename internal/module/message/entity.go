package message

import "time"

// MessageDirection represents the direction of a message
type MessageDirection string

const (
	DirectionToPatient   MessageDirection = "to_patient"
	DirectionFromPatient MessageDirection = "from_patient"
	DirectionStaffToStaff MessageDirection = "staff_to_staff"
)

// Message represents a message in the system
type Message struct {
	ID              string           `json:"id" db:"id"`
	RoomID          string           `json:"room_id" db:"room_id"`
	PatientID       *string          `json:"patient_id,omitempty" db:"patient_id"`
	SenderStaffID   *string          `json:"sender_staff_id,omitempty" db:"sender_staff_id"`
	ReceiverStaffID *string          `json:"receiver_staff_id,omitempty" db:"receiver_staff_id"`
	Direction       MessageDirection `json:"direction" db:"direction"`
	Content         string           `json:"content" db:"content"`
	IsRead          bool             `json:"is_read" db:"is_read"`
	IsUrgent        bool             `json:"is_urgent" db:"is_urgent"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	ReadAt          *time.Time       `json:"read_at,omitempty" db:"read_at"`
}

// MessageWithDetails includes additional info for display
type MessageWithDetails struct {
	Message
	RoomName        string  `json:"room_name,omitempty"`
	PatientName     string  `json:"patient_name,omitempty"`
	SenderName      string  `json:"sender_name,omitempty"`
	ReceiverName    string  `json:"receiver_name,omitempty"`
}
