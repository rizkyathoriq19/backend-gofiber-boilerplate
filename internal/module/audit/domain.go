package audit

import "context"

// AuditRepository defines the interface for audit log data operations
type AuditRepository interface {
	Create(ctx context.Context, entry *AuditLogEntry) error
	GetByID(ctx context.Context, id string) (*AuditLog, error)
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error)
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
}

// AuditUseCase defines the interface for audit log business logic
type AuditUseCase interface {
	Log(ctx context.Context, entry *AuditLogEntry) error
	GetByID(ctx context.Context, id string) (*AuditLog, error)
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error)
	Cleanup(ctx context.Context, retentionDays int) (int64, error)
}
