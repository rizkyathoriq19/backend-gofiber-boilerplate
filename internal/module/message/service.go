package message

import (
	"boilerplate-be/internal/shared/errors"
)

type messageUseCase struct {
	messageRepo MessageRepository
}

// NewMessageUseCase creates a new message use case
func NewMessageUseCase(messageRepo MessageRepository) MessageUseCase {
	return &messageUseCase{
		messageRepo: messageRepo,
	}
}

// SendMessage sends a new message
func (u *messageUseCase) SendMessage(req *SendMessageRequest, senderID string, isStaff bool) (*Message, error) {
	msg := &Message{
		RoomID:    req.RoomID,
		PatientID: req.PatientID,
		Direction: req.Direction,
		Content:   req.Content,
		IsUrgent:  req.IsUrgent,
	}

	// Set sender based on who is sending
	if isStaff {
		msg.SenderStaffID = &senderID
		msg.ReceiverStaffID = req.ReceiverStaffID
	} else {
		// Patient is sending - set patient_id
		msg.PatientID = &senderID
	}

	if err := u.messageRepo.Create(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// GetMessage gets a message by ID
func (u *messageUseCase) GetMessage(id string) (*Message, error) {
	msg, err := u.messageRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// GetRoomMessages gets all messages for a room
func (u *messageUseCase) GetRoomMessages(roomID string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	return u.messageRepo.GetByRoomID(roomID, filter)
}

// GetPatientMessages gets all messages for a patient
func (u *messageUseCase) GetPatientMessages(patientID string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	return u.messageRepo.GetByPatientID(patientID, filter)
}

// GetStaffMessages gets all messages for/from a staff member
func (u *messageUseCase) GetStaffMessages(staffID string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	return u.messageRepo.GetByStaffID(staffID, filter)
}

// GetUnreadCount gets unread message count for a user
func (u *messageUseCase) GetUnreadCount(userID string, isStaff bool) (int, error) {
	return u.messageRepo.GetUnreadCount(userID, isStaff)
}

// MarkAsRead marks a message as read
func (u *messageUseCase) MarkAsRead(id string) error {
	// Verify message exists
	_, err := u.messageRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.messageRepo.MarkAsRead(id)
}

// MarkAllAsRead marks all messages in a room as read for a user
func (u *messageUseCase) MarkAllAsRead(roomID string, userID string) error {
	return u.messageRepo.MarkAllAsRead(roomID, userID)
}

// DeleteMessage deletes a message
func (u *messageUseCase) DeleteMessage(id string) error {
	// Verify message exists
	msg, err := u.messageRepo.GetByID(id)
	if err != nil {
		return err
	}

	if msg == nil {
		return errors.New(errors.ResourceNotFound)
	}

	return u.messageRepo.Delete(id)
}
