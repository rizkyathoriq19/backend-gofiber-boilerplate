package audit

import (
	"context"
)

type auditUseCase struct {
	repo AuditRepository
}

// NewAuditUseCase creates a new audit use case
func NewAuditUseCase(repo AuditRepository) AuditUseCase {
	return &auditUseCase{repo: repo}
}

func (uc *auditUseCase) Log(ctx context.Context, entry *AuditLogEntry) error {
	return uc.repo.Create(ctx, entry)
}

func (uc *auditUseCase) GetByID(ctx context.Context, id string) (*AuditLog, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *auditUseCase) List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error) {
	return uc.repo.List(ctx, filter)
}

func (uc *auditUseCase) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays < 30 {
		retentionDays = 30 // Minimum 30 days retention
	}
	return uc.repo.DeleteOlderThan(ctx, retentionDays)
}
