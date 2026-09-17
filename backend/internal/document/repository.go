package document

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ================================================================== */
/*  Column list — keep in ONE place so every query stays consistent    */
/* ================================================================== */

// documentColumns is the exact SELECT list used by every read query.
// The DB column `public_id` holds a Cloudinary public_id and maps to the
// Go field `Document.PublicID`.
const documentColumns = `
	id, folder_id, user_id, name, public_id, mime_type, size,
	description, version, status, comment, uploaded_at
`

/* ================================================================== */
/*  Insert                                                             */
/* ================================================================== */

// InsertDocumentTx inserts a document inside an existing transaction and
// fills doc.ID and doc.UploadedAt from the RETURNING clause.
func InsertDocumentTx(ctx context.Context, tx pgx.Tx, doc *Document) error {
	query := `
		INSERT INTO documents
			(folder_id, user_id, name, public_id, mime_type, size,
			 description, version, status, comment, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING id, uploaded_at;
	`
	return tx.QueryRow(
		ctx,
		query,
		doc.FolderID,
		doc.UserID,
		doc.Name,
		doc.PublicID, // Cloudinary public_id
		doc.MimeType,
		doc.Size,
		doc.Description,
		doc.Version,
		doc.Status,
		doc.Comment,
	).Scan(&doc.ID, &doc.UploadedAt)
}

// InsertDocument is the convenience version when you don't already have a
// transaction open. Internally it opens one and delegates to InsertDocumentTx.
func InsertDocument(ctx context.Context, db *pgxpool.Pool, doc *Document) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := InsertDocumentTx(ctx, tx, doc); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

/* ================================================================== */
/*  Read                                                               */
/* ================================================================== */

// scanDocument reads a row in the exact order of `documentColumns`.
func scanDocument(row pgx.Row) (*Document, error) {
	var d Document
	err := row.Scan(
		&d.ID,
		&d.FolderID,
		&d.UserID,
		&d.Name,
		&d.PublicID, // Cloudinary public_id
		&d.MimeType,
		&d.Size,
		&d.Description,
		&d.Version,
		&d.Status,
		&d.Comment,
		&d.UploadedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// GetDocumentByID returns a single document. Returns (nil, nil) when not found
// so callers can distinguish "not found" from "db error".
func GetDocumentByID(ctx context.Context, db *pgxpool.Pool, docID int64) (*Document, error) {
	query := `SELECT ` + documentColumns + ` FROM documents WHERE id = $1`
	doc, err := scanDocument(db.QueryRow(ctx, query, docID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get document by id: %w", err)
	}
	return doc, nil
}

// GetDocumentsByFolder returns all documents inside a folder, newest first.
func GetDocumentsByFolder(ctx context.Context, db *pgxpool.Pool, folderID int64) ([]Document, error) {
	query := `SELECT ` + documentColumns + `
		FROM documents
		WHERE folder_id = $1
		ORDER BY uploaded_at DESC`

	rows, err := db.Query(ctx, query, folderID)
	if err != nil {
		return nil, fmt.Errorf("query documents by folder: %w", err)
	}
	defer rows.Close()

	docs := make([]Document, 0)
	for rows.Next() {
		var d Document
		if err := rows.Scan(
			&d.ID,
			&d.FolderID,
			&d.UserID,
			&d.Name,
			&d.PublicID,
			&d.MimeType,
			&d.Size,
			&d.Description,
			&d.Version,
			&d.Status,
			&d.Comment,
			&d.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

// SearchDocuments performs a case-insensitive search on name + description.
// If folderID is nil, the search runs across all folders.
func SearchDocuments(
	ctx context.Context,
	db *pgxpool.Pool,
	term string,
	folderID *int64,
) ([]Document, error) {
	pattern := "%" + term + "%"

	var (
		query string
		args  []any
	)

	if folderID != nil {
		query = `SELECT ` + documentColumns + `
			FROM documents
			WHERE folder_id = $1
			  AND (name ILIKE $2 OR description ILIKE $2)
			ORDER BY uploaded_at DESC`
		args = []any{*folderID, pattern}
	} else {
		query = `SELECT ` + documentColumns + `
			FROM documents
			WHERE name ILIKE $1 OR description ILIKE $1
			ORDER BY uploaded_at DESC`
		args = []any{pattern}
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search documents: %w", err)
	}
	defer rows.Close()

	docs := make([]Document, 0)
	for rows.Next() {
		var d Document
		if err := rows.Scan(
			&d.ID,
			&d.FolderID,
			&d.UserID,
			&d.Name,
			&d.PublicID,
			&d.MimeType,
			&d.Size,
			&d.Description,
			&d.Version,
			&d.Status,
			&d.Comment,
			&d.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

/* ================================================================== */
/*  Update                                                             */
/* ================================================================== */

// UpdateDocumentName changes the display name of a document.
func UpdateDocumentName(ctx context.Context, db *pgxpool.Pool, docID int64, newName string) error {
	tag, err := db.Exec(ctx,
		`UPDATE documents SET name = $1 WHERE id = $2`,
		newName, docID,
	)
	if err != nil {
		return fmt.Errorf("update document name: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

// MoveDocument moves a document to a different folder.
func MoveDocument(ctx context.Context, db *pgxpool.Pool, docID, newFolderID int64) error {
	tag, err := db.Exec(ctx,
		`UPDATE documents SET folder_id = $1 WHERE id = $2`,
		newFolderID, docID,
	)
	if err != nil {
		return fmt.Errorf("move document: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

// UpdateDocumentStatus updates the workflow status (and optional comment)
// and returns the folder_id the document belongs to — useful for the
// service to check folder-level permissions afterwards.
func UpdateDocumentStatus(
	ctx context.Context,
	db *pgxpool.Pool,
	docID int64,
	newStatus string,
	comment string,
) (int64, error) {
	var folderID int64
	query := `
		UPDATE documents
		SET status = $1, comment = $2
		WHERE id = $3
		RETURNING folder_id;
	`
	err := db.QueryRow(ctx, query, newStatus, comment, docID).Scan(&folderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("document not found")
		}
		return 0, fmt.Errorf("update document status: %w", err)
	}
	return folderID, nil
}

// UpdateDocumentPointer re-points the master row at a new Cloudinary
// public_id and version number. Used both by UpdateDocumentVersion and by
// RestoreVersion.
func UpdateDocumentPointer(
	ctx context.Context,
	db *pgxpool.Pool,
	docID int64,
	version int,
	publicID string,
) error {
	tag, err := db.Exec(ctx,
		`UPDATE documents SET version = $1, public_id = $2 WHERE id = $3`,
		version, publicID, docID,
	)
	if err != nil {
		return fmt.Errorf("update document pointer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

/* ================================================================== */
/*  Delete                                                             */
/* ================================================================== */

// DeleteDocumentRecord removes the master row. Version rows should be
// deleted by the caller (usually inside the same transaction).
func DeleteDocumentRecord(ctx context.Context, db *pgxpool.Pool, docID int64) error {
	tag, err := db.Exec(ctx, `DELETE FROM documents WHERE id = $1`, docID)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

// DeleteDocumentTx is the transaction-aware version used during the
// two-step delete (versions first, then master).
func DeleteDocumentTx(ctx context.Context, tx pgx.Tx, docID int64) error {
	tag, err := tx.Exec(ctx, `DELETE FROM documents WHERE id = $1`, docID)
	if err != nil {
		return fmt.Errorf("delete document (tx): %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

/* ================================================================== */
/*  Helpers — cross-table lookups used by the status workflow          */
/* ================================================================== */

// FatchUserFromFOlder returns the owner (user_id) of a folder.
// (Kept the original misspelled name so call sites don't break.)
func FatchUserFromFOlder(ctx context.Context, db *pgxpool.Pool, folderID int64) (int64, error) {
	var userID int64
	err := db.QueryRow(ctx,
		`SELECT created_by FROM folders WHERE id = $1`, folderID,
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("folder not found")
		}
		return 0, fmt.Errorf("fetch folder owner: %w", err)
	}
	return userID, nil
}
