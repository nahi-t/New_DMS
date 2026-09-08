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

type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// 2. Accept DBTX instead of *pgxpool.Pool
func (r *Repository) InsertDocumentVersion(ctx context.Context, db DBTX, docVersion *DocumentVersion, hashedString string) error {
	query := `
        INSERT INTO document_versions (user_id, document_id, version, file_path, hashed_string)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id;
    `
	return db.QueryRow(
		ctx,
		query,
		docVersion.UserID,
		docVersion.DocumentID,
		docVersion.Version,
		docVersion.FilePath,
		hashedString,
	).Scan(&docVersion.ID)

}

func getDocumentVersionsByDocumentID(ctx context.Context, db *pgxpool.Pool, documentID int64) ([]DocumentVersion, error) {
	query := `
		SELECT id, user_id, document_id, version, file_path, hashed_string
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
		if err := rows.Scan(&v.ID, &v.UserID, &v.DocumentID, &v.Version, &v.FilePath, &v.HashedString); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func getDocumentVersionByID(ctx context.Context, db *pgxpool.Pool, id int64) (*DocumentVersion, error) {
	query := `
		SELECT id, user_id, document_id, version, file_path, hashed_string
		FROM document_versions
		WHERE id = $1
	`
	var v DocumentVersion
	err := db.QueryRow(ctx, query, id).Scan(&v.ID, &v.UserID, &v.DocumentID, &v.Version, &v.FilePath)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func getAllDocumentVersions(ctx context.Context, db *pgxpool.Pool) ([]DocumentVersion, error) {
	query := `
		SELECT id, user_id, document_id, version, file_path
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
		if err := rows.Scan(&v.ID, &v.UserID, &v.DocumentID, &v.Version, &v.FilePath); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}
