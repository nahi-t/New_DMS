// internal/auth/audit.go
package auth

import "context"

type AuditLogger interface {
	Log(ctx context.Context, event string) error
	LogLogin(
		ctx context.Context,
		userID *int64,
		userName string,
		event string,
		ip string,
		userAgent string,
		success bool,
	) error
}
