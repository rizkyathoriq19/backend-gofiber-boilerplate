package audit

import "time"

// AuditLogListRequest represents the request for listing audit logs
type AuditLogListRequest struct {
	UserID       string     `query:"user_id"`
	Action       string     `query:"action"`
	ResourceType string     `query:"resource_type"`
	ResourceID   string     `query:"resource_id"`
	StartDate    *time.Time `query:"start_date"`
	EndDate      *time.Time `query:"end_date"`
	Page         int        `query:"page"`
	PageSize     int        `query:"page_size"`
}

// ToFilter converts request to filter
func (r *AuditLogListRequest) ToFilter() AuditLogFilter {
	return AuditLogFilter{
		UserID:       r.UserID,
		Action:       Action(r.Action),
		ResourceType: r.ResourceType,
		ResourceID:   r.ResourceID,
		StartDate:    r.StartDate,
		EndDate:      r.EndDate,
		Page:         r.Page,
		PageSize:     r.PageSize,
	}
}
