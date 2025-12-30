package middleware

import (
	"boilerplate-be/internal/module/audit"

	"github.com/gofiber/fiber/v2"
)

// AuditContext stores audit information in Fiber context
type AuditContext struct {
	UserID    *string
	IPAddress string
	UserAgent string
	RequestID string
}

// GetAuditContext retrieves audit context from Fiber context
func GetAuditContext(c *fiber.Ctx) *AuditContext {
	userID, _ := c.Locals("userID").(string)
	requestID, _ := c.Locals("requestID").(string)

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	return &AuditContext{
		UserID:    userIDPtr,
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
		RequestID: requestID,
	}
}

// CreateAuditEntry creates an audit log entry from Fiber context
func CreateAuditEntry(c *fiber.Ctx, action audit.Action, resourceType, resourceID string, oldValues, newValues map[string]interface{}) *audit.AuditLogEntry {
	ctx := GetAuditContext(c)

	return &audit.AuditLogEntry{
		UserID:       ctx.UserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValues:    oldValues,
		NewValues:    newValues,
		IPAddress:    ctx.IPAddress,
		UserAgent:    ctx.UserAgent,
		RequestID:    ctx.RequestID,
	}
}

// AuditLogger is a helper for logging audit events
type AuditLogger struct {
	useCase audit.AuditUseCase
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(useCase audit.AuditUseCase) *AuditLogger {
	return &AuditLogger{useCase: useCase}
}

// Log creates an audit log entry
func (al *AuditLogger) Log(c *fiber.Ctx, action audit.Action, resourceType, resourceID string, oldValues, newValues map[string]interface{}) error {
	entry := CreateAuditEntry(c, action, resourceType, resourceID, oldValues, newValues)
	return al.useCase.Log(c.Context(), entry)
}

// LogCreate logs a CREATE action
func (al *AuditLogger) LogCreate(c *fiber.Ctx, resourceType, resourceID string, newValues map[string]interface{}) error {
	return al.Log(c, audit.ActionCreate, resourceType, resourceID, nil, newValues)
}

// LogUpdate logs an UPDATE action
func (al *AuditLogger) LogUpdate(c *fiber.Ctx, resourceType, resourceID string, oldValues, newValues map[string]interface{}) error {
	return al.Log(c, audit.ActionUpdate, resourceType, resourceID, oldValues, newValues)
}

// LogDelete logs a DELETE action
func (al *AuditLogger) LogDelete(c *fiber.Ctx, resourceType, resourceID string, oldValues map[string]interface{}) error {
	return al.Log(c, audit.ActionDelete, resourceType, resourceID, oldValues, nil)
}

// LogLogin logs a LOGIN action
func (al *AuditLogger) LogLogin(c *fiber.Ctx, userID string, metadata map[string]interface{}) error {
	entry := CreateAuditEntry(c, audit.ActionLogin, "auth", userID, nil, nil)
	entry.Metadata = metadata
	return al.useCase.Log(c.Context(), entry)
}

// LogLogout logs a LOGOUT action
func (al *AuditLogger) LogLogout(c *fiber.Ctx, userID string) error {
	return al.Log(c, audit.ActionLogout, "auth", userID, nil, nil)
}
