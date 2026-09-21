package auditlog

import (
	"context"
	"errors"
	"fmt"

	"github.com/docmanage_new/internal/auth"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Log extracts userID and username from ctx and writes the audit event.
func (s *Service) Log(ctx context.Context, event string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	userName, err := auth.UsernameFromContext(ctx)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	return s.repo.InsertAuditLog(ctx, userID, userName, event)
}

// LogTx is the transactional variant.
func (s *Service) LogTx(ctx context.Context, tx pgx.Tx, event string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	userName, err := auth.UsernameFromContext(ctx)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}
	return s.repo.InsertAuditLogTx(ctx, tx, userID, userName, event)
}

// LogAsSystem is for background jobs, cron, migrations — no request context.
func (s *Service) LogAsSystem(ctx context.Context, event string) error {
	const systemUserID int64 = 0
	const systemUserName = "system"
	return s.repo.InsertAuditLog(ctx, systemUserID, systemUserName, event)
}

var ErrNoUser = errors.New("no user in context")

func (s *Service) GetAuditLogs(
	ctx context.Context,
	f AuditFilter,
) ([]AuditLog, int, error) {
	// Clamp values defensively — never trust the handler alone.
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 500 {
		f.Limit = 500
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return s.repo.GetAuditLogs(ctx, f)
}
