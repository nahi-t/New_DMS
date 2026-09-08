package documentversion

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Repo *Repository // Changed from DB to Repo to match what you use below
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

func CreateDocumentVersion(ctx context.Context, db *pgxpool.Pool, docVersion *DocumentVersion, file io.ReadSeeker) error { // 1. Calculate the hash
	hash, err := CalculateHash(file)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}
	docVersion.HashedString = hash

	// 2. Rewind the file stream back to start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to rewind file: %w", err)
	}

	// 3. Begin Transaction
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 4. Insert using the transaction
	query := `
        INSERT INTO document_versions (user_id, document_id, version, file_path, hashed_string)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id;
    `
	err = tx.QueryRow(
		ctx,
		query,
		docVersion.UserID,
		docVersion.DocumentID,
		docVersion.Version,
		docVersion.FilePath,
		docVersion.HashedString,
	).Scan(&docVersion.ID)

	if err != nil {
		return fmt.Errorf("failed to insert document version: %w", err)
	}

	// 5. Commit Transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UpdateDocumentVersion(ctx context.Context, db *pgxpool.Pool, docID int64, file multipart.File, header *multipart.FileHeader, userID int64) error {
	// 1. Read file bytes to calculate hash and save to disk later
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	newHash, err := CalculateHash(bytes.NewReader(fileBytes))
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	// 2. Begin Transaction
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Lock the document row to prevent concurrent version increments
	//    Cast version to integer in case it's stored as text
	var currentDocVersion int
	lockQuery := `SELECT version::INT FROM documents WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, lockQuery, docID).Scan(&currentDocVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("document with id %d does not exist", docID)
		}
		return fmt.Errorf("failed to lock document: %w", err)
	}

	// 4. Fetch the latest version & hash for this document from document_versions
	var currentVersion int
	var currentHash string
	queryFetch := `
		SELECT version::INT, hashed_string 
		FROM document_versions 
		WHERE document_id = $1 
		ORDER BY version DESC 
		LIMIT 1
	`
	err = tx.QueryRow(ctx, queryFetch, docID).Scan(&currentVersion, &currentHash)
	hasRows := true
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			hasRows = false // First version (no prior versions)
		} else {
			return fmt.Errorf("failed to fetch latest version: %w", err)
		}
	}

	// 5. Compare hashes – if identical, skip and rollback (no change)
	if hasRows && currentHash == newHash {
		return nil // defer will rollback and release lock
	}

	// 6. Determine new version number
	newVersion := 1
	if hasRows {
		newVersion = currentVersion + 1
	}

	// 7. Save the new file to disk
	uploadDir := "./uploads_withNewVersion"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}
	newFileName := fmt.Sprintf("doc_%d_v%d_%s", docID, newVersion, header.Filename)
	newFilePath := filepath.Join(uploadDir, newFileName)

	if err := os.WriteFile(newFilePath, fileBytes, 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// 8. Insert new record into document_versions (version is stored as integer, but if your column is varchar, it will be auto-converted)
	queryInsert := `
		INSERT INTO document_versions (document_id, version, user_id, file_path, hashed_string)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, queryInsert, docID, newVersion, userID, newFilePath, newHash)
	if err != nil {
		return fmt.Errorf("failed to insert document version: %w", err)
	}

	// 9. Update the version column in the main documents table (FIXED – now uncommented)
	//    If the column is text, the integer will be accepted; if you want to be explicit, use $1::TEXT
	queryUpdateMain := `UPDATE documents SET version = $1 WHERE id = $2`
	_, err = tx.Exec(ctx, queryUpdateMain, newVersion, docID)
	if err != nil {
		return fmt.Errorf("failed to update main document version: %w", err)
	}

	// 10. Commit Transaction
	return tx.Commit(ctx)
}
func (s *Service) GetDocumentVersionsByDocumentID(ctx context.Context, db *pgxpool.Pool, documentID int64) ([]DocumentVersion, error) {
	versions, err := getDocumentVersionsByDocumentID(ctx, db, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document versions: %w", err)
	}
	return versions, nil
}
