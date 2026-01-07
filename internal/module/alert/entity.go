package alert

import (
	"time"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeVoiceCall   AlertType = "voice_call"
	AlertTypeButtonPress AlertType = "button_press"
	AlertTypeEmergency   AlertType = "emergency"
	AlertTypeSystem      AlertType = "system"
	AlertTypeScheduled   AlertType = "scheduled"
)

// AlertPriority represents the priority level of an alert
type AlertPriority string

const (
	PriorityCritical AlertPriority = "critical"
	PriorityHigh     AlertPriority = "high"
	PriorityMedium   AlertPriority = "medium"
	PriorityLow      AlertPriority = "low"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	StatusPending      AlertStatus = "pending"
	StatusAcknowledged AlertStatus = "acknowledged"
	StatusInProgress   AlertStatus = "in_progress"
	StatusResolved     AlertStatus = "resolved"
	StatusEscalated    AlertStatus = "escalated"
	StatusCancelled    AlertStatus = "cancelled"
)

// Alert represents a nurse call alert
type Alert struct {
	ID                       string        `json:"id" db:"id"`
	RoomID                   string        `json:"room_id" db:"room_id"`
	PatientID                *string       `json:"patient_id" db:"patient_id"`
	DeviceID                 *string       `json:"device_id" db:"device_id"`
	AssignedStaffID          *string       `json:"assigned_staff_id" db:"assigned_staff_id"`
	ResolvedByStaffID        *string       `json:"resolved_by_staff_id" db:"resolved_by_staff_id"`
	Type                     AlertType     `json:"type" db:"type"`
	Priority                 AlertPriority `json:"priority" db:"priority"`
	Status                   AlertStatus   `json:"status" db:"status"`
	Message                  string        `json:"message" db:"message"`
	DetectedKeywords         []string      `json:"detected_keywords" db:"detected_keywords"`
	AudioReference           string        `json:"audio_reference" db:"audio_reference"`
	EscalationCount          int           `json:"escalation_count" db:"escalation_count"`
	EscalationTimeoutMinutes int           `json:"escalation_timeout_minutes" db:"escalation_timeout_minutes"`
	CreatedAt                time.Time     `json:"created_at" db:"created_at"`
	AcknowledgedAt           *time.Time    `json:"acknowledged_at" db:"acknowledged_at"`
	ResolvedAt               *time.Time    `json:"resolved_at" db:"resolved_at"`
	UpdatedAt                time.Time     `json:"updated_at" db:"updated_at"`
}

// AlertWithDetails includes related entity details
type AlertWithDetails struct {
	Alert
	RoomName      string `json:"room_name"`
	PatientName   string `json:"patient_name,omitempty"`
	AssignedStaff string `json:"assigned_staff,omitempty"`
}

// AlertHistory represents a history entry for an alert
type AlertHistory struct {
	ID             string      `json:"id" db:"id"`
	AlertID        string      `json:"alert_id" db:"alert_id"`
	StaffID        *string     `json:"staff_id" db:"staff_id"`
	Action         string      `json:"action" db:"action"`
	PreviousStatus AlertStatus `json:"previous_status" db:"previous_status"`
	NewStatus      AlertStatus `json:"new_status" db:"new_status"`
	Notes          string      `json:"notes" db:"notes"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
}

// AlertFilter represents filters for querying alerts
type AlertFilter struct {
	RoomID    string        `query:"room_id"`
	PatientID string        `query:"patient_id"`
	StaffID   string        `query:"staff_id"`
	Type      AlertType     `query:"type"`
	Priority  AlertPriority `query:"priority"`
	Status    AlertStatus   `query:"status"`
	Page      int           `query:"page"`
	Limit     int           `query:"limit"`
}
