package share

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

/* ------------------------------------------------------------------ */
/*  Reads                                                              */
/* ------------------------------------------------------------------ */

// GetFolderOwner returns the user_id that created the folder.
func (r *Repository) GetFolderOwner(ctx context.Context, folderID int64) (int64, error) {
	var ownerID int64
	err := r.pool.QueryRow(ctx,
		`SELECT created_by FROM folders WHERE id = $1`, folderID,
	).Scan(&ownerID)
	if err != nil {
		return 0, fmt.Errorf("get folder owner: %w", err)
	}
	return ownerID, nil
}

// GetDocumentFolderID returns the folder_id for a document.
func (r *Repository) GetDocumentFolderID(ctx context.Context, documentID int64) (int64, error) {
	var folderID int64
	err := r.pool.QueryRow(ctx,
		`SELECT folder_id FROM documents WHERE id = $1`, documentID,
	).Scan(&folderID)
	if err != nil {
		return 0, fmt.Errorf("get document folder: %w", err)
	}
	return folderID, nil
}

// GetDocumentOwner returns the user_id that uploaded the document.
func (r *Repository) GetDocumentOwner(ctx context.Context, documentID int64) (int64, error) {
	var ownerID int64
	err := r.pool.QueryRow(ctx,
		`SELECT created_by FROM documents WHERE id = $1`, documentID,
	).Scan(&ownerID)
	if err != nil {
		return 0, fmt.Errorf("get document owner: %w", err)
	}
	return ownerID, nil
}

// ListShares returns everyone a document is shared with.
func (r *Repository) ListShares(ctx context.Context, documentID int64) ([]DocumentShare, error) {
	const q = `
		SELECT ds.id, ds.document_id,
		       d.name AS document_name,
		       d.description AS document_description,
		       d.folder_id,
		       d.version,
		       ds.user_id, u.username AS user_name, u.email AS user_email,
		       ds.permission, ds.shared_by,
		       ub.username AS shared_by_name,
		       ds.shared_at
		FROM document_shares ds
		JOIN documents d ON d.id = ds.document_id
		JOIN users u   ON u.id = ds.user_id
		JOIN users ub  ON ub.id = ds.shared_by
		WHERE ds.document_id = $1
		ORDER BY ds.shared_at DESC
	`
	rows, err := r.pool.Query(ctx, q, documentID)
	if err != nil {
		return nil, fmt.Errorf("list shares: %w", err)
	}
	defer rows.Close()
	return scanShares(rows)
}

// GetSharedWithMe returns every document shared with a user.
func (r *Repository) GetSharedWithMe(ctx context.Context, userID int64) ([]DocumentShare, error) {
	const q = `
		SELECT ds.id, ds.document_id,
		       d.name AS document_name,
		       d.description AS document_description,
		       d.folder_id,
		       d.version,
		       ds.user_id, u.username AS user_name, u.email AS user_email,
		       ds.permission, ds.shared_by,
		       ub.username AS shared_by_name,
		       ds.shared_at
		FROM document_shares ds
		JOIN documents d ON d.id = ds.document_id
		JOIN users u   ON u.id = ds.user_id
		JOIN users ub  ON ub.id = ds.shared_by
		WHERE ds.user_id = $1
		ORDER BY ds.shared_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("get shared with me: %w", err)
	}
	defer rows.Close()
	return scanShares(rows)
}

// CanAccess reports whether the user has at least the required permission.
// "editor" implies "viewer".
func (r *Repository) CanAccess(
	ctx context.Context,
	documentID, userID int64,
	required string,
) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM document_shares
			WHERE document_id = $1
			  AND user_id     = $2
			  AND (permission = $3 OR permission = 'editor')
		)
	`
	var ok bool
	err := r.pool.QueryRow(ctx, q, documentID, userID, required).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("can access: %w", err)
	}
	return ok, nil
}

/* ------------------------------------------------------------------ */
/*  Writes                                                             */
/* ------------------------------------------------------------------ */

// Share grants (or updates) a user's access to a document.
func (r *Repository) Share(
	ctx context.Context,
	documentID, userID, sharedBy int64,
	permission string,
) error {
	const q = `
		INSERT INTO document_shares
			(document_id, user_id, permission, shared_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (document_id, user_id)
		DO UPDATE SET permission = EXCLUDED.permission,
		              shared_by  = EXCLUDED.shared_by,
		              shared_at  = now()
	`
	if _, err := r.pool.Exec(ctx, q, documentID, userID, permission, sharedBy); err != nil {
		return fmt.Errorf("share document: %w", err)
	}
	return nil
}

// Unshare removes a user's access to a document.
func (r *Repository) Unshare(ctx context.Context, documentID, userID int64) error {
	const q = `DELETE FROM document_shares WHERE document_id = $1 AND user_id = $2`
	if _, err := r.pool.Exec(ctx, q, documentID, userID); err != nil {
		return fmt.Errorf("unshare document: %w", err)
	}
	return nil
}

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanShares(rows rowScanner) ([]DocumentShare, error) {
	var out []DocumentShare
	for rows.Next() {
		var s DocumentShare
		if err := rows.Scan(
			&s.ID,
			&s.DocumentID,
			&s.DocumentName,
			&s.DocumentDescription,
			&s.FolderID,
			&s.Version,
			&s.UserID,
			&s.UserName,
			&s.UserEmail,
			&s.Permission,
			&s.SharedBy,
			&s.SharedByName,
			&s.SharedAt,
		); err != nil {
			return nil, fmt.Errorf("scan share: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate shares: %w", err)
	}
	return out, nil
}
