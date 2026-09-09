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
		doc.MimeType, doc.Size, doc.Description, doc.Version,
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

// UpdateDocumentName updates the document's name.
func UpdateDocumentName(ctx context.Context, db *pgxpool.Pool, docID int64, newName string) error {
	query := `UPDATE documents SET name = $1 WHERE id = $2`
	res, err := db.Exec(ctx, query, newName, docID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

// MoveDocument changes the folder of a document.
func MoveDocument(ctx context.Context, db *pgxpool.Pool, docID, newFolderID int64) error {
	// Optional: check that newFolderID exists (you can add a foreign key constraint)
	query := `UPDATE documents SET folder_id = $1 WHERE id = $2`
	res, err := db.Exec(ctx, query, newFolderID, docID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

// SearchDocuments returns documents matching a search term, optionally filtered by folder.
func SearchDocuments(ctx context.Context, db *pgxpool.Pool, searchTerm string, folderID *int64) ([]Document, error) {
	var query string
	var args []interface{}
	if folderID != nil {
		query = `SELECT id, folder_id, user_id, name, file_path, mime_type, size, uploaded_at, description, version
                 FROM documents
                 WHERE name ILIKE $1 AND folder_id = $2
                 ORDER BY uploaded_at DESC`
		args = append(args, "%"+searchTerm+"%", *folderID)
	} else {
		query = `SELECT id, folder_id, user_id, name, file_path, mime_type, size, uploaded_at, description, version
                 FROM documents
                 WHERE name ILIKE $1
                 ORDER BY uploaded_at DESC`
		args = append(args, "%"+searchTerm+"%")
	}
	rows, err := db.Query(ctx, query, args...)
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
func UpdateDocumentStatus(
	ctx context.Context,
	db *pgxpool.Pool,
	docID int64,
	status string,
	comment string,
) (int64, error) {

	query := `
		UPDATE documents
		SET status = $1, comment = $2
		WHERE id = $3
		RETURNING folder_id
	`

	var folderID int64

	err := db.QueryRow(ctx, query, status, comment, docID).Scan(&folderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("document not found")
		}
		return 0, err
	}

	return folderID, nil
}
func FatchUserFromFOlder(ctx context.Context, db *pgxpool.Pool, folderID int64) (int64, error) {
	query := `
		SELECT created_by
		FROM folders
		WHERE id = $1
	`

	var userID int64

	err := db.QueryRow(ctx, query, folderID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("folder not found")
		}
		return 0, err
	}

	return userID, nil
}
