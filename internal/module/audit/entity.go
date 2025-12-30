package audit

import (
	"time"
)

// Action represents the type of action performed
type Action string

const (
	ActionCreate Action = "CREATE"
	ActionRead   Action = "READ"
	ActionUpdate Action = "UPDATE"
	ActionDelete Action = "DELETE"
	ActionLogin  Action = "LOGIN"
	ActionLogout Action = "LOGOUT"
	ActionExport Action = "EXPORT"
	ActionImport Action = "IMPORT"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       *string                `json:"user_id,omitempty"`
	Action       Action                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	OldValues    map[string]interface{} `json:"old_values,omitempty"`
	NewValues    map[string]interface{} `json:"new_values,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// AuditLogFilter for querying audit logs
type AuditLogFilter struct {
	UserID       string
	Action       Action
	ResourceType string
	ResourceID   string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}

// AuditLogEntry is used to create new audit log entries
type AuditLogEntry struct {
	UserID       *string
	Action       Action
	ResourceType string
	ResourceID   string
	OldValues    map[string]interface{}
	NewValues    map[string]interface{}
	IPAddress    string
	UserAgent    string
	RequestID    string
	Metadata     map[string]interface{}
}
