// internal/auditLog/model.go (or wherever AuditLog is defined)
package auditlog

import "time"

type AuditLog struct {
	ID                int64     `db:"id"                  json:"id"`
	UserID            int64     `db:"user_id"             json:"user_id"`
	UserName          string    `db:"user_name"           json:"user_name"`
	Event             string    `db:"event"               json:"event"`
	EventHappenedTime time.Time `db:"event_happened_time" json:"event_happened_time"`
	CreatedAt         time.Time `db:"created_at"          json:"created_at"`
}
