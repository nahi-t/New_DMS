package document

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertDocument(ctx context.Context, db *pgxpool.Pool, doc *Document) error {
	query := `
        INSERT INTO documents (folder_id, user_id, name, file_path, mime_type, size, description, version)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, uploaded_at
    `
	return db.QueryRow(ctx, query,
		doc.FolderID, doc.UserID, doc.Name, doc.FilePath,
		doc.MimeType, doc.Size, doc.Description, 1,
	).Scan(&doc.ID, &doc.UploadedAt)
}

func GetDocumentsByFolder(ctx context.Context, db *pgxpool.Pool, folderID int64) ([]Document, error) {
	query := `
        SELECT id, folder_id, user_id, name, file_path, mime_type, size, uploaded_at, description, version
        FROM documents
        WHERE folder_id = $1
        ORDER BY uploaded_at DESC
    `
	rows, err := db.Query(ctx, query, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.FolderID, &d.UserID, &d.Name, &d.FilePath,
			&d.MimeType, &d.Size, &d.UploadedAt, &d.Description, &d.Version); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func GetDocumentByID(ctx context.Context, db *pgxpool.Pool, id int64) (*Document, error) {
	query := `
        SELECT id, folder_id, user_id, name, file_path, mime_type, size, uploaded_at, description, version
        FROM documents
        WHERE id = $1
    `
	var d Document
	err := db.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.FolderID, &d.UserID, &d.Name, &d.FilePath,
		&d.MimeType, &d.Size, &d.UploadedAt, &d.Description, &d.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("document not found")
	}
	return &d, err
}

// DeleteDocumentRecord removes the document record from DB.
func DeleteDocumentRecord(ctx context.Context, db *pgxpool.Pool, id int64) error {
	query := `DELETE FROM documents WHERE id = $1`
	res, err := db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}
