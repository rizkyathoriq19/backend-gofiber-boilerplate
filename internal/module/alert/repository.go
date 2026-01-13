package alert

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"boilerplate-be/internal/shared/errors"
	"boilerplate-be/internal/shared/utils"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type alertRepository struct {
	db          *sql.DB
	cacheHelper *utils.CacheHelper
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *sql.DB, cacheHelper *utils.CacheHelper) AlertRepository {
	return &alertRepository{
		db:          db,
		cacheHelper: cacheHelper,
	}
}

// Create creates a new alert
func (r *alertRepository) Create(alert *Alert) error {
	id, _ := uuid.NewV7()
	alert.ID = id.String()
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	alert.Status = StatusPending
	if alert.EscalationTimeoutMinutes == 0 {
		alert.EscalationTimeoutMinutes = 5 // Default 5 minutes
	}

	query := `
		INSERT INTO alerts (id, room_id, patient_id, device_id, assigned_staff_id, type, priority, status, message, detected_keywords, audio_reference, escalation_count, escalation_timeout_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(query, alert.ID, alert.RoomID, alert.PatientID, alert.DeviceID, alert.AssignedStaffID, alert.Type, alert.Priority, alert.Status, alert.Message, pq.Array(alert.DetectedKeywords), alert.AudioReference, alert.EscalationCount, alert.EscalationTimeoutMinutes, alert.CreatedAt, alert.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

// GetByID gets an alert by ID
func (r *alertRepository) GetByID(id string) (*Alert, error) {
	alert := &Alert{}
	query := `
		SELECT id, room_id, patient_id, device_id, assigned_staff_id, resolved_by_staff_id, type, priority, status, message, detected_keywords, audio_reference, escalation_count, escalation_timeout_minutes, created_at, acknowledged_at, resolved_at, updated_at
		FROM alerts
		WHERE id = $1
	`
	var keywords pq.StringArray
	err := r.db.QueryRow(query, id).Scan(
		&alert.ID, &alert.RoomID, &alert.PatientID, &alert.DeviceID, &alert.AssignedStaffID, &alert.ResolvedByStaffID,
		&alert.Type, &alert.Priority, &alert.Status, &alert.Message, &keywords, &alert.AudioReference,
		&alert.EscalationCount, &alert.EscalationTimeoutMinutes, &alert.CreatedAt, &alert.AcknowledgedAt, &alert.ResolvedAt, &alert.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	alert.DetectedKeywords = keywords
	return alert, nil
}

// GetAll gets all alerts with filters
func (r *alertRepository) GetAll(filter *AlertFilter) ([]*AlertWithDetails, int, error) {
	var alerts []*AlertWithDetails
	var total int

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}

	baseQuery := `FROM alerts a 
		LEFT JOIN rooms rm ON a.room_id = rm.id 
		LEFT JOIN patients p ON a.patient_id = p.id 
		LEFT JOIN staff s ON a.assigned_staff_id = s.id 
		LEFT JOIN users u ON s.user_id = u.id
		WHERE 1=1`
	var args []interface{}
	argIndex := 1

	if filter.RoomID != "" {
		baseQuery += fmt.Sprintf(" AND a.room_id = $%d", argIndex)
		args = append(args, filter.RoomID)
		argIndex++
	}
	if filter.PatientID != "" {
		baseQuery += fmt.Sprintf(" AND a.patient_id = $%d", argIndex)
		args = append(args, filter.PatientID)
		argIndex++
	}
	if filter.StaffID != "" {
		baseQuery += fmt.Sprintf(" AND a.assigned_staff_id = $%d", argIndex)
		args = append(args, filter.StaffID)
		argIndex++
	}
	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND a.type = $%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}
	if filter.Priority != "" {
		baseQuery += fmt.Sprintf(" AND a.priority = $%d", argIndex)
		args = append(args, filter.Priority)
		argIndex++
	}
	if filter.Status != "" {
		baseQuery += fmt.Sprintf(" AND a.status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	countQuery := `SELECT COUNT(*) ` + baseQuery
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	offset := (filter.Page - 1) * filter.Limit
	// Order by priority (critical first), then by created_at
	selectQuery := fmt.Sprintf(`SELECT a.id, a.room_id, a.patient_id, a.device_id, a.assigned_staff_id, a.resolved_by_staff_id, a.type, a.priority, a.status, a.message, a.detected_keywords, a.audio_reference, a.escalation_count, a.escalation_timeout_minutes, a.created_at, a.acknowledged_at, a.resolved_at, a.updated_at, COALESCE(rm.name, ''), COALESCE(p.name, ''), COALESCE(u.name, '') %s ORDER BY CASE a.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, a.created_at DESC LIMIT $%d OFFSET $%d`, baseQuery, argIndex, argIndex+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		alert := &AlertWithDetails{}
		var keywords pq.StringArray
		err := rows.Scan(
			&alert.ID, &alert.RoomID, &alert.PatientID, &alert.DeviceID, &alert.AssignedStaffID, &alert.ResolvedByStaffID,
			&alert.Type, &alert.Priority, &alert.Status, &alert.Message, &keywords, &alert.AudioReference,
			&alert.EscalationCount, &alert.EscalationTimeoutMinutes, &alert.CreatedAt, &alert.AcknowledgedAt, &alert.ResolvedAt, &alert.UpdatedAt,
			&alert.RoomName, &alert.PatientName, &alert.AssignedStaff,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		alert.DetectedKeywords = keywords
		alerts = append(alerts, alert)
	}

	return alerts, total, nil
}

// GetActiveAlerts gets all active (non-resolved) alerts sorted by priority
func (r *alertRepository) GetActiveAlerts() ([]*AlertWithDetails, error) {
	var alerts []*AlertWithDetails

	query := `
		SELECT a.id, a.room_id, a.patient_id, a.device_id, a.assigned_staff_id, a.resolved_by_staff_id, a.type, a.priority, a.status, a.message, a.detected_keywords, a.audio_reference, a.escalation_count, a.escalation_timeout_minutes, a.created_at, a.acknowledged_at, a.resolved_at, a.updated_at, COALESCE(rm.name, ''), COALESCE(p.name, ''), COALESCE(u.name, '')
		FROM alerts a 
		LEFT JOIN rooms rm ON a.room_id = rm.id 
		LEFT JOIN patients p ON a.patient_id = p.id 
		LEFT JOIN staff s ON a.assigned_staff_id = s.id 
		LEFT JOIN users u ON s.user_id = u.id
		WHERE a.status NOT IN ('resolved', 'cancelled')
		ORDER BY CASE a.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, a.created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		alert := &AlertWithDetails{}
		var keywords pq.StringArray
		err := rows.Scan(
			&alert.ID, &alert.RoomID, &alert.PatientID, &alert.DeviceID, &alert.AssignedStaffID, &alert.ResolvedByStaffID,
			&alert.Type, &alert.Priority, &alert.Status, &alert.Message, &keywords, &alert.AudioReference,
			&alert.EscalationCount, &alert.EscalationTimeoutMinutes, &alert.CreatedAt, &alert.AcknowledgedAt, &alert.ResolvedAt, &alert.UpdatedAt,
			&alert.RoomName, &alert.PatientName, &alert.AssignedStaff,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		alert.DetectedKeywords = keywords
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetActiveAlertsByRoom gets active alerts for a specific room
func (r *alertRepository) GetActiveAlertsByRoom(roomID string) ([]*AlertWithDetails, error) {
	var alerts []*AlertWithDetails

	query := `
		SELECT a.id, a.room_id, a.patient_id, a.device_id, a.assigned_staff_id, a.resolved_by_staff_id, a.type, a.priority, a.status, a.message, a.detected_keywords, a.audio_reference, a.escalation_count, a.escalation_timeout_minutes, a.created_at, a.acknowledged_at, a.resolved_at, a.updated_at, COALESCE(rm.name, ''), COALESCE(p.name, ''), COALESCE(u.name, '')
		FROM alerts a 
		LEFT JOIN rooms rm ON a.room_id = rm.id 
		LEFT JOIN patients p ON a.patient_id = p.id 
		LEFT JOIN staff s ON a.assigned_staff_id = s.id 
		LEFT JOIN users u ON s.user_id = u.id
		WHERE a.room_id = $1 AND a.status NOT IN ('resolved', 'cancelled')
		ORDER BY CASE a.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, a.created_at ASC
	`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		alert := &AlertWithDetails{}
		var keywords pq.StringArray
		err := rows.Scan(
			&alert.ID, &alert.RoomID, &alert.PatientID, &alert.DeviceID, &alert.AssignedStaffID, &alert.ResolvedByStaffID,
			&alert.Type, &alert.Priority, &alert.Status, &alert.Message, &keywords, &alert.AudioReference,
			&alert.EscalationCount, &alert.EscalationTimeoutMinutes, &alert.CreatedAt, &alert.AcknowledgedAt, &alert.ResolvedAt, &alert.UpdatedAt,
			&alert.RoomName, &alert.PatientName, &alert.AssignedStaff,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		alert.DetectedKeywords = keywords
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetPendingAlerts gets all pending alerts
func (r *alertRepository) GetPendingAlerts() ([]*Alert, error) {
	var alerts []*Alert

	query := `
		SELECT id, room_id, patient_id, device_id, assigned_staff_id, resolved_by_staff_id, type, priority, status, message, detected_keywords, audio_reference, escalation_count, escalation_timeout_minutes, created_at, acknowledged_at, resolved_at, updated_at
		FROM alerts
		WHERE status = 'pending'
		ORDER BY CASE priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END, created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		alert := &Alert{}
		var keywords pq.StringArray
		err := rows.Scan(
			&alert.ID, &alert.RoomID, &alert.PatientID, &alert.DeviceID, &alert.AssignedStaffID, &alert.ResolvedByStaffID,
			&alert.Type, &alert.Priority, &alert.Status, &alert.Message, &keywords, &alert.AudioReference,
			&alert.EscalationCount, &alert.EscalationTimeoutMinutes, &alert.CreatedAt, &alert.AcknowledgedAt, &alert.ResolvedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		alert.DetectedKeywords = keywords
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetPendingAlertsForEscalation gets pending alerts that need escalation
func (r *alertRepository) GetPendingAlertsForEscalation(timeoutMinutes int) ([]*Alert, error) {
	var alerts []*Alert

	query := `
		SELECT id, room_id, patient_id, device_id, assigned_staff_id, resolved_by_staff_id, type, priority, status, message, detected_keywords, audio_reference, escalation_count, escalation_timeout_minutes, created_at, acknowledged_at, resolved_at, updated_at
		FROM alerts
		WHERE status = 'pending' 
		AND created_at < NOW() - INTERVAL '1 minute' * $1
		ORDER BY priority, created_at ASC
	`

	rows, err := r.db.Query(query, timeoutMinutes)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		alert := &Alert{}
		var keywords pq.StringArray
		err := rows.Scan(
			&alert.ID, &alert.RoomID, &alert.PatientID, &alert.DeviceID, &alert.AssignedStaffID, &alert.ResolvedByStaffID,
			&alert.Type, &alert.Priority, &alert.Status, &alert.Message, &keywords, &alert.AudioReference,
			&alert.EscalationCount, &alert.EscalationTimeoutMinutes, &alert.CreatedAt, &alert.AcknowledgedAt, &alert.ResolvedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		alert.DetectedKeywords = keywords
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// Update updates an alert
func (r *alertRepository) Update(alert *Alert) error {
	alert.UpdatedAt = time.Now()

	query := `
		UPDATE alerts 
		SET status = $2, assigned_staff_id = $3, resolved_by_staff_id = $4, message = $5, escalation_count = $6, acknowledged_at = $7, resolved_at = $8, updated_at = $9
		WHERE id = $1
	`

	result, err := r.db.Exec(query, alert.ID, alert.Status, alert.AssignedStaffID, alert.ResolvedByStaffID, alert.Message, alert.EscalationCount, alert.AcknowledgedAt, alert.ResolvedAt, alert.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// UpdateStatus updates the alert status
func (r *alertRepository) UpdateStatus(id string, status AlertStatus, staffID *string) error {
	query := `UPDATE alerts SET status = $2, assigned_staff_id = COALESCE($3, assigned_staff_id), updated_at = $4 WHERE id = $1`

	result, err := r.db.Exec(query, id, status, staffID, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// Acknowledge acknowledges an alert
func (r *alertRepository) Acknowledge(id, staffID string) error {
	now := time.Now()
	query := `UPDATE alerts SET status = $2, assigned_staff_id = $3, acknowledged_at = $4, updated_at = $5 WHERE id = $1`

	result, err := r.db.Exec(query, id, StatusAcknowledged, staffID, now, now)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// Resolve resolves an alert
func (r *alertRepository) Resolve(id, staffID string) error {
	now := time.Now()
	query := `UPDATE alerts SET status = $2, resolved_by_staff_id = $3, resolved_at = $4, updated_at = $5 WHERE id = $1`

	result, err := r.db.Exec(query, id, StatusResolved, staffID, now, now)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// Escalate escalates an alert
func (r *alertRepository) Escalate(id string, newStaffID *string) error {
	query := `UPDATE alerts SET status = $2, assigned_staff_id = $3, escalation_count = escalation_count + 1, updated_at = $4 WHERE id = $1`

	result, err := r.db.Exec(query, id, StatusEscalated, newStaffID, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

// Delete deletes an alert
func (r *alertRepository) Delete(id string) error {
	query := `DELETE FROM alerts WHERE id = $1`

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

// CreateHistory creates an alert history entry
func (r *alertRepository) CreateHistory(history *AlertHistory) error {
	id, _ := uuid.NewV7()
	history.ID = id.String()
	history.CreatedAt = time.Now()

	query := `
		INSERT INTO alert_history (id, alert_id, staff_id, action, previous_status, new_status, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(query, history.ID, history.AlertID, history.StaffID, history.Action, history.PreviousStatus, history.NewStatus, history.Notes, history.CreatedAt)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	_ = r.cacheHelper.Delete(context.Background(), fmt.Sprintf("alert:history:%s", history.AlertID))

	return nil
}

// GetHistory gets alert history
func (r *alertRepository) GetHistory(alertID string) ([]*AlertHistory, error) {
	var history []*AlertHistory

	query := `
		SELECT id, alert_id, staff_id, action, previous_status, new_status, notes, created_at
		FROM alert_history
		WHERE alert_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, alertID)
	if err != nil {
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	for rows.Next() {
		h := &AlertHistory{}
		err := rows.Scan(&h.ID, &h.AlertID, &h.StaffID, &h.Action, &h.PreviousStatus, &h.NewStatus, &h.Notes, &h.CreatedAt)
		if err != nil {
			return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
		}
		history = append(history, h)
	}

	return history, nil
}
