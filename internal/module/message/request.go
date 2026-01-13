package message

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	RoomID          string           `json:"room_id" validate:"required,uuid"`
	PatientID       *string          `json:"patient_id" validate:"omitempty,uuid"`
	ReceiverStaffID *string          `json:"receiver_staff_id" validate:"omitempty,uuid"`
	Direction       MessageDirection `json:"direction" validate:"required,oneof=to_patient from_patient staff_to_staff"`
	Content         string           `json:"content" validate:"required,min=1,max=2000"`
	IsUrgent        bool             `json:"is_urgent"`
}

// MessageFilter represents filters for message queries
type MessageFilter struct {
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
	Direction string `query:"direction"`
	IsRead    *bool  `query:"is_read"`
	IsUrgent  *bool  `query:"is_urgent"`
}
