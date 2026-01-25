package alert

import (
	"time"
)

// AlertResponse represents the alert response
type AlertResponse struct {
	ID                       string        `json:"uuid"`
	RoomID                   string        `json:"room_uuid"`
	PatientID                *string       `json:"patient_uuid"`
	DeviceID                 *string       `json:"device_uuid"`
	AssignedStaffID          *string       `json:"assigned_staff_uuid"`
	ResolvedByStaffID        *string       `json:"resolved_by_staff_uuid"`
	Type                     AlertType     `json:"type"`
	Priority                 AlertPriority `json:"priority"`
	Status                   AlertStatus   `json:"status"`
	Message                  string        `json:"message"`
	DetectedKeywords         []string      `json:"detected_keywords,omitempty"`
	EscalationCount          int           `json:"escalation_count"`
	EscalationTimeoutMinutes int           `json:"escalation_timeout_minutes"`
	CreatedAt                time.Time     `json:"created_at"`
	AcknowledgedAt           *time.Time    `json:"acknowledged_at,omitempty"`
	ResolvedAt               *time.Time    `json:"resolved_at,omitempty"`
	UpdatedAt                time.Time     `json:"updated_at"`
}

// AlertWithDetailsResponse includes related entity details
type AlertWithDetailsResponse struct {
	AlertResponse
	RoomName      string `json:"room_name"`
	PatientName   string `json:"patient_name,omitempty"`
	AssignedStaff string `json:"assigned_staff,omitempty"`
}

// AlertListResponse represents paginated alert list
type AlertListResponse struct {
	Alerts     []*AlertWithDetailsResponse `json:"alerts"`
	Total      int                         `json:"total"`
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalPages int                         `json:"total_pages"`
}

// AlertHistoryResponse represents a history entry
type AlertHistoryResponse struct {
	ID             string      `json:"uuid"`
	Action         string      `json:"action"`
	PreviousStatus AlertStatus `json:"previous_status"`
	NewStatus      AlertStatus `json:"new_status"`
	Notes          string      `json:"notes,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

// ToResponse converts Alert entity to AlertResponse
func (a *Alert) ToResponse() *AlertResponse {
	return &AlertResponse{
		ID:                       a.ID,
		RoomID:                   a.RoomID,
		PatientID:                a.PatientID,
		DeviceID:                 a.DeviceID,
		AssignedStaffID:          a.AssignedStaffID,
		ResolvedByStaffID:        a.ResolvedByStaffID,
		Type:                     a.Type,
		Priority:                 a.Priority,
		Status:                   a.Status,
		Message:                  a.Message,
		DetectedKeywords:         a.DetectedKeywords,
		EscalationCount:          a.EscalationCount,
		EscalationTimeoutMinutes: a.EscalationTimeoutMinutes,
		CreatedAt:                a.CreatedAt,
		AcknowledgedAt:           a.AcknowledgedAt,
		ResolvedAt:               a.ResolvedAt,
		UpdatedAt:                a.UpdatedAt,
	}
}

// ToResponse converts AlertWithDetails to AlertWithDetailsResponse
func (a *AlertWithDetails) ToResponse() *AlertWithDetailsResponse {
	return &AlertWithDetailsResponse{
		AlertResponse: *a.Alert.ToResponse(),
		RoomName:      a.RoomName,
		PatientName:   a.PatientName,
		AssignedStaff: a.AssignedStaff,
	}
}
