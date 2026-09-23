// internal/auditLog/model.go (or wherever AuditLog is defined)
package auditlog

import "time"

type AuditLog struct {
	ID                int64     `db:"id"                  json:"id"                  example:"42"`
	UserID            *int64    `db:"user_id"             json:"user_id"             example:"1"`
	UserName          string    `db:"user_name"           json:"user_name"           example:"admin"`
	Event             string    `db:"event"               json:"event"               example:"document.download"`
	EventHappenedTime time.Time `db:"event_happened_time" json:"event_happened_time" example:"2026-09-21T14:48:32Z"`
	CreatedAt         time.Time `db:"created_at"          json:"created_at"          example:"2026-09-21T14:48:32Z"`
	IPAddress         *string   `db:"ip_address"          json:"ip_address,omitempty"  example:"203.0.113.10"`
	UserAgent         *string   `db:"user_agent"          json:"user_agent,omitempty"  example:"Mozilla/5.0"`
	Success           bool      `db:"success"             json:"success"             example:"true"`
}

type AuditLogPage struct {
	Data   []AuditLog `json:"data"`
	Total  int        `json:"total"  example:"128"`
	Limit  int        `json:"limit"  example:"50"`
	Offset int        `json:"offset" example:"0"`
}
