package share

import (
	"context"
	"time"
)

// DocumentShare represents a single share row returned to the frontend.
type DocumentShare struct {
	ID                  int64   `db:"id"                  json:"id"`
	DocumentID          int64   `db:"document_id"         json:"document_id"`
	DocumentName        string  `db:"document_name"       json:"document_name"`
	DocumentDescription *string `db:"document_description" json:"document_description,omitempty"`
	FolderID            int64   `db:"folder_id"           json:"folder_id"`
	Version             int     `db:"version"             json:"version"`

	UserID    int64  `db:"user_id"    json:"user_id"`
	UserName  string `db:"user_name"  json:"user_name"`
	UserEmail string `db:"user_email" json:"user_email"`

	Permission   string    `db:"permission"     json:"permission"`
	SharedBy     int64     `db:"shared_by"      json:"shared_by"`
	SharedByName string    `db:"shared_by_name" json:"shared_by_name"`
	SharedAt     time.Time `db:"shared_at"      json:"shared_at"`
}

// AuditEntry mirrors the auditlog insert used by the auth package.
// Kept local so this package does not depend on audit internals.
type AuditEntry struct {
	UserID    *int64
	UserName  string
	Event     string
	IP        string
	UserAgent string
	Success   bool
}

// AuditLogger is the minimal interface this package needs from auditlog.Service.
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
