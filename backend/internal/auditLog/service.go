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

/* ------------------------------------------------------------------ */
/*  Context-based (user already authenticated)                         */
/* ------------------------------------------------------------------ */

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
	if userName == "" {
		userName = fmt.Sprintf("user_%d", userID)
	}

	return s.repo.InsertAuditEntry(ctx, AuditEntry{
		UserID:   &userID,
		UserName: userName,
		Event:    event,
		Success:  true,
	})
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
	if userName == "" {
		userName = fmt.Sprintf("user_%d", userID)
	}

	if err := s.repo.InsertAuditLogTx(ctx, tx, userID, userName, event); err != nil {
		return err
	}
	return nil
}

// LogAsSystem is for background jobs, cron, migrations — no request context.
func (s *Service) LogAsSystem(ctx context.Context, event string) error {
	const systemUserName = "system"
	return s.repo.InsertAuditEntry(ctx, AuditEntry{
		UserID:   nil, // system has no user row
		UserName: systemUserName,
		Event:    event,
		Success:  true,
	})
}

/* ------------------------------------------------------------------ */
/*  Login / logout (no user in context yet)                            */
/* ------------------------------------------------------------------ */

// LogLogin records a login attempt.
// Pass nil for userID when the login failed and no user was identified.
func (s *Service) LogLogin(
	ctx context.Context,
	userID *int64,
	userName string,
	event string,
	ip string,
	userAgent string,
	success bool,
) error {
	if userName == "" {
		userName = "unknown"
	}
	return s.repo.InsertAuditEntry(ctx, AuditEntry{
		UserID:    userID,
		UserName:  userName,
		Event:     event,
		IP:        ip,
		UserAgent: userAgent,
		Success:   success,
	})
}

/* ------------------------------------------------------------------ */
/*  Query                                                              */
/* ------------------------------------------------------------------ */

var ErrNoUser = errors.New("no user in context")

func (s *Service) GetAuditLogs(
	ctx context.Context,
	f AuditFilter,
) ([]AuditLog, int, error) {
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
