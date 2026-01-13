package message

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/utils"

	"github.com/google/uuid"
)

type messageRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB, cacheHelper *utils.CacheHelper) MessageRepository {
	return &messageRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// Create creates a new message
func (r *messageRepository) Create(msg *Message) error {
	id, _ := uuid.NewV7()
	msg.ID = id.String()
	msg.CreatedAt = time.Now()
	msg.IsRead = false

	query := `
		INSERT INTO messages (id, room_id, patient_id, sender_staff_id, receiver_staff_id, direction, content, is_read, is_urgent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query, msg.ID, msg.RoomID, msg.PatientID, msg.SenderStaffID, msg.ReceiverStaffID, msg.Direction, msg.Content, msg.IsRead, msg.IsUrgent, msg.CreatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetByID gets a message by ID
func (r *messageRepository) GetByID(id string) (*Message, error) {
	msg := &Message{}
	query := `
		SELECT id, room_id, patient_id, sender_staff_id, receiver_staff_id, direction, content, is_read, is_urgent, created_at, read_at
		FROM messages
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&msg.ID, &msg.RoomID, &msg.PatientID, &msg.SenderStaffID, &msg.ReceiverStaffID,
		&msg.Direction, &msg.Content, &msg.IsRead, &msg.IsUrgent, &msg.CreatedAt, &msg.ReadAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	return msg, nil
}

// GetByRoomID gets messages by room
func (r *messageRepository) GetByRoomID(roomID string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	return r.getMessages("room_id", roomID, filter)
}

// GetByPatientID gets messages for a patient
func (r *messageRepository) GetByPatientID(patientID string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	return r.getMessages("patient_id", patientID, filter)
}

// GetByStaffID gets messages sent to or from a staff member
func (r *messageRepository) GetByStaffID(staffID string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}

	var messages []*MessageWithDetails
	var total int

	baseQuery := `FROM messages m 
		LEFT JOIN rooms rm ON m.room_id = rm.id
		LEFT JOIN patients p ON m.patient_id = p.id
		WHERE (m.sender_staff_id = $1 OR m.receiver_staff_id = $1)`
	args := []interface{}{staffID}
	argIndex := 2

	if filter.Direction != "" {
		baseQuery += fmt.Sprintf(" AND m.direction = $%d", argIndex)
		args = append(args, filter.Direction)
		argIndex++
	}
	if filter.IsRead != nil {
		baseQuery += fmt.Sprintf(" AND m.is_read = $%d", argIndex)
		args = append(args, *filter.IsRead)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`SELECT m.id, m.room_id, m.patient_id, m.sender_staff_id, m.receiver_staff_id, m.direction, m.content, m.is_read, m.is_urgent, m.created_at, m.read_at, COALESCE(rm.name, ''), COALESCE(p.name, '') %s ORDER BY m.created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		msg := &MessageWithDetails{}
		err := rows.Scan(
			&msg.ID, &msg.RoomID, &msg.PatientID, &msg.SenderStaffID, &msg.ReceiverStaffID,
			&msg.Direction, &msg.Content, &msg.IsRead, &msg.IsUrgent, &msg.CreatedAt, &msg.ReadAt,
			&msg.RoomName, &msg.PatientName,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

func (r *messageRepository) getMessages(field string, value string, filter *MessageFilter) ([]*MessageWithDetails, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}

	var messages []*MessageWithDetails
	var total int

	baseQuery := fmt.Sprintf(`FROM messages m 
		LEFT JOIN rooms rm ON m.room_id = rm.id
		LEFT JOIN patients p ON m.patient_id = p.id
		WHERE m.%s = $1`, field)
	args := []interface{}{value}
	argIndex := 2

	if filter.Direction != "" {
		baseQuery += fmt.Sprintf(" AND m.direction = $%d", argIndex)
		args = append(args, filter.Direction)
		argIndex++
	}
	if filter.IsRead != nil {
		baseQuery += fmt.Sprintf(" AND m.is_read = $%d", argIndex)
		args = append(args, *filter.IsRead)
		argIndex++
	}
	if filter.IsUrgent != nil {
		baseQuery += fmt.Sprintf(" AND m.is_urgent = $%d", argIndex)
		args = append(args, *filter.IsUrgent)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`SELECT m.id, m.room_id, m.patient_id, m.sender_staff_id, m.receiver_staff_id, m.direction, m.content, m.is_read, m.is_urgent, m.created_at, m.read_at, COALESCE(rm.name, ''), COALESCE(p.name, '') %s ORDER BY m.created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		msg := &MessageWithDetails{}
		err := rows.Scan(
			&msg.ID, &msg.RoomID, &msg.PatientID, &msg.SenderStaffID, &msg.ReceiverStaffID,
			&msg.Direction, &msg.Content, &msg.IsRead, &msg.IsUrgent, &msg.CreatedAt, &msg.ReadAt,
			&msg.RoomName, &msg.PatientName,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		messages = append(messages, msg)
	}

	return messages, total, nil
}

// GetUnreadCount gets unread message count
func (r *messageRepository) GetUnreadCount(userID string, isStaff bool) (int, error) {
	var count int
	var query string

	if isStaff {
		query = `SELECT COUNT(*) FROM messages WHERE (receiver_staff_id = $1 OR (sender_staff_id IS NULL AND direction = 'from_patient')) AND is_read = false`
	} else {
		query = `SELECT COUNT(*) FROM messages WHERE patient_id = $1 AND direction = 'to_patient' AND is_read = false`
	}

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	return count, nil
}

// MarkAsRead marks a message as read
func (r *messageRepository) MarkAsRead(id string) error {
	query := `UPDATE messages SET is_read = true, read_at = $2 WHERE id = $1`

	result, err := r.db.Exec(query, id, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// MarkAllAsRead marks all messages in a room as read
func (r *messageRepository) MarkAllAsRead(roomID string, userID string) error {
	query := `UPDATE messages SET is_read = true, read_at = $3 WHERE room_id = $1 AND (receiver_staff_id = $2 OR patient_id = $2) AND is_read = false`

	_, err := r.db.Exec(query, roomID, userID, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	_ = r.cacheHelper.Delete(context.Background(), fmt.Sprintf("messages:room:%s", roomID))

	return nil
}

// Delete deletes a message
func (r *messageRepository) Delete(id string) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}
