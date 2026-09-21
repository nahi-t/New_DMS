package auditlog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) InsertAuditLog(
	ctx context.Context,
	userID int64,
	userName string,
	event string,
) error {
	const q = `
        INSERT INTO auditlog (user_id, user_name, event, event_happened_time)
        VALUES ($1, $2, $3, now())
    `
	_, err := r.pool.Exec(ctx, q, userID, userName, event)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (r *Repository) InsertAuditLogTx(
	ctx context.Context,
	tx pgx.Tx,
	userID int64,
	userName string,
	event string,
) error {
	const q = `
        INSERT INTO auditlog (user_id, user_name, event, event_happened_time)
        VALUES ($1, $2, $3, now())
    `
	_, err := tx.Exec(ctx, q, userID, userName, event)
	if err != nil {
		return fmt.Errorf("insert audit log tx: %w", err)
	}
	return nil
}

type AuditFilter struct {
	Limit  int
	Offset int
	UserID string
	Event  string
	From   *time.Time
	To     *time.Time
}

func (r *Repository) GetAuditLogs(
	ctx context.Context,
	f AuditFilter,
) ([]AuditLog, int, error) {
	// ---- Build WHERE clause safely with numbered placeholders ----
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if f.UserID != "" {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, f.UserID)
		i++
	}
	if f.Event != "" {
		where = append(where, fmt.Sprintf("event = $%d", i))
		args = append(args, f.Event)
		i++
	}
	if f.From != nil {
		where = append(where, fmt.Sprintf("event_happened_time >= $%d", i))
		args = append(args, *f.From)
		i++
	}
	if f.To != nil {
		where = append(where, fmt.Sprintf("event_happened_time <= $%d", i))
		args = append(args, *f.To)
		i++
	}

	clause := "WHERE " + strings.Join(where, " AND ")

	// ---- Count total (for pagination UI) ----
	var total int
	countQ := "SELECT COUNT(*) FROM auditlog " + clause
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	// ---- Fetch page ----
	args = append(args, f.Limit, f.Offset)
	listQ := fmt.Sprintf(`
		SELECT id, user_id, user_name, event, event_happened_time, created_at
		FROM auditlog
		%s
		ORDER BY event_happened_time DESC
		LIMIT $%d OFFSET $%d
	`, clause, i, i+1)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get audit logs: %w", err)
	}
	defer rows.Close()

	logs := make([]AuditLog, 0, f.Limit)
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.UserName,
			&l.Event,
			&l.EventHappenedTime,
			&l.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit logs: %w", err)
	}

	return logs, total, nil
}
