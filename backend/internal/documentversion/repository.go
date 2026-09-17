package documentversion

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

// DBTX lets the Repository accept either *pgxpool.Pool or pgx.Tx.
type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

/* ------------------------------------------------------------------ */
/*  Insert                                                             */
/* ------------------------------------------------------------------ */

// InsertDocumentVersion inserts a new row and fills docVersion.ID and
// docVersion.CreatedAt from the DB.
func (r *Repository) InsertDocumentVersion(
	ctx context.Context,
	db DBTX,
	docVersion *DocumentVersion,
	hashedString string,
) error {
	query := `
        INSERT INTO document_versions
            (user_id, document_id, version, public_id, original_filename, hashed_string)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at;
    `
	return db.QueryRow(
		ctx,
		query,
		docVersion.UserID,
		docVersion.DocumentID,
		docVersion.Version,
		docVersion.PublicID,
		docVersion.OriginalFilename,
		hashedString,
	).Scan(&docVersion.ID, &docVersion.CreatedAt)
}

/* ------------------------------------------------------------------ */
/*  Fetch                                                              */
/* ------------------------------------------------------------------ */

// getDocumentVersionsByDocumentID returns all versions for one document,
// newest first.
func getDocumentVersionsByDocumentID(
	ctx context.Context,
	db *pgxpool.Pool,
	documentID int64,
) ([]DocumentVersion, error) {
	query := `
		SELECT id, user_id, document_id, version, public_id,
		       original_filename, hashed_string, created_at
		FROM document_versions
		WHERE document_id = $1
		ORDER BY version DESC
	`
	rows, err := db.Query(ctx, query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []DocumentVersion
	for rows.Next() {
		var v DocumentVersion
		if err := rows.Scan(
			&v.ID,
			&v.UserID,
			&v.DocumentID,
			&v.Version,
			&v.PublicID,
			&v.OriginalFilename,
			&v.HashedString,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// getDocumentVersionByID returns a single version by its UUID.
// NOTE: ID is a string (UUID), not int64.
func getDocumentVersionByID(
	ctx context.Context,
	db *pgxpool.Pool,
	id string,
) (*DocumentVersion, error) {
	query := `
		SELECT id, user_id, document_id, version, public_id,
		       original_filename, hashed_string, created_at
		FROM document_versions
		WHERE id = $1
	`
	var v DocumentVersion
	err := db.QueryRow(ctx, query, id).Scan(
		&v.ID,
		&v.UserID,
		&v.DocumentID,
		&v.Version,
		&v.PublicID,
		&v.OriginalFilename,
		&v.HashedString,
		&v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// getAllDocumentVersions returns every version in the system,
// grouped by document then newest version first.
func getAllDocumentVersions(
	ctx context.Context,
	db *pgxpool.Pool,
) ([]DocumentVersion, error) {
	query := `
		SELECT id, user_id, document_id, version, public_id,
		       original_filename, hashed_string, created_at
		FROM document_versions
		ORDER BY document_id, version DESC
	`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []DocumentVersion
	for rows.Next() {
		var v DocumentVersion
		if err := rows.Scan(
			&v.ID,
			&v.UserID,
			&v.DocumentID,
			&v.Version,
			&v.PublicID,
			&v.OriginalFilename,
			&v.HashedString,
			&v.CreatedAt,
		); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
