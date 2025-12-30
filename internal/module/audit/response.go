package audit

// AuditLogResponse represents the response for an audit log entry
type AuditLogResponse struct {
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
	CreatedAt    string                 `json:"created_at"`
}

// ToResponse converts AuditLog to AuditLogResponse
func (al *AuditLog) ToResponse() *AuditLogResponse {
	return &AuditLogResponse{
		ID:           al.ID,
		UserID:       al.UserID,
		Action:       al.Action,
		ResourceType: al.ResourceType,
		ResourceID:   al.ResourceID,
		OldValues:    al.OldValues,
		NewValues:    al.NewValues,
		IPAddress:    al.IPAddress,
		UserAgent:    al.UserAgent,
		RequestID:    al.RequestID,
		Metadata:     al.Metadata,
		CreatedAt:    al.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToResponseList converts slice of AuditLog to slice of AuditLogResponse
func ToResponseList(logs []AuditLog) []AuditLogResponse {
	result := make([]AuditLogResponse, len(logs))
	for i, log := range logs {
		result[i] = *log.ToResponse()
	}
	return result
}
