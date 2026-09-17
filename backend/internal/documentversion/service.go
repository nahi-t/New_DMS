// package documentversion

// import (
// 	"bytes"
// 	"context"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"mime/multipart"
// 	"os"
// 	"path/filepath"

// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// type Service struct {
// 	Repo *Repository // Changed from DB to Repo to match what you use below
// }

// func NewService(repo *Repository) *Service {
// 	return &Service{Repo: repo}
// }

// func CreateDocumentVersion(ctx context.Context, db *pgxpool.Pool, docVersion *DocumentVersion, file io.ReadSeeker) error { // 1. Calculate the hash
// 	hash, err := CalculateHash(file)
// 	if err != nil {
// 		return fmt.Errorf("failed to calculate hash: %w", err)
// 	}
// 	docVersion.HashedString = hash

// 	// 2. Rewind the file stream back to start
// 	if _, err := file.Seek(0, io.SeekStart); err != nil {
// 		return fmt.Errorf("failed to rewind file: %w", err)
// 	}

// 	// 3. Begin Transaction
// 	tx, err := db.Begin(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback(ctx)

// 	// 4. Insert using the transaction
// 	query := `
//         INSERT INTO document_versions (user_id, document_id, version, file_path, hashed_string)
//         VALUES ($1, $2, $3, $4, $5)
//         RETURNING id;
//     `
// 	err = tx.QueryRow(
// 		ctx,
// 		query,
// 		docVersion.UserID,
// 		docVersion.DocumentID,
// 		docVersion.Version,
// 		docVersion.FilePath,
// 		docVersion.HashedString,
// 	).Scan(&docVersion.ID)

// 	if err != nil {
// 		return fmt.Errorf("failed to insert document version: %w", err)
// 	}

// 	// 5. Commit Transaction
// 	if err := tx.Commit(ctx); err != nil {
// 		return fmt.Errorf("failed to commit transaction: %w", err)
// 	}

// 	return nil
// }

// func UpdateDocumentVersion(ctx context.Context, db *pgxpool.Pool, docID int64, file multipart.File, header *multipart.FileHeader, userID int64) error {
// 	// 1. Read file bytes to calculate hash and save to disk later
// 	fileBytes, err := io.ReadAll(file)
// 	if err != nil {
// 		return fmt.Errorf("failed to read file: %w", err)
// 	}

// 	newHash, err := CalculateHash(bytes.NewReader(fileBytes))
// 	if err != nil {
// 		return fmt.Errorf("failed to calculate hash: %w", err)
// 	}

// 	// 2. Begin Transaction
// 	tx, err := db.Begin(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback(ctx)

// 	// 3. Lock the document row to prevent concurrent version increments
// 	//    Cast version to integer in case it's stored as text
// 	var currentDocVersion int
// 	lockQuery := `SELECT version::INT FROM documents WHERE id = $1 FOR UPDATE`
// 	err = tx.QueryRow(ctx, lockQuery, docID).Scan(&currentDocVersion)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return fmt.Errorf("document with id %d does not exist", docID)
// 		}
// 		return fmt.Errorf("failed to lock document: %w", err)
// 	}

// 	// 4. Fetch the latest version & hash for this document from document_versions
// 	var currentVersion int
// 	var currentHash string
// 	queryFetch := `
// 		SELECT version::INT, hashed_string
// 		FROM document_versions
// 		WHERE document_id = $1
// 		ORDER BY version DESC
// 		LIMIT 1
// 	`
// 	err = tx.QueryRow(ctx, queryFetch, docID).Scan(&currentVersion, &currentHash)
// 	hasRows := true
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			hasRows = false // First version (no prior versions)
// 		} else {
// 			return fmt.Errorf("failed to fetch latest version: %w", err)
// 		}
// 	}

// 	// 5. Compare hashes – if identical, skip and rollback (no change)
// 	if hasRows && currentHash == newHash {
// 		return nil // defer will rollback and release lock
// 	}

// 	// 6. Determine new version number
// 	newVersion := 1
// 	if hasRows {
// 		newVersion = currentVersion + 1
// 	}

// 	// 7. Save the new file to disk
// 	uploadDir := "./uploads_withNewVersion"
// 	if err := os.MkdirAll(uploadDir, 0755); err != nil {
// 		return fmt.Errorf("failed to create upload directory: %w", err)
// 	}
// 	newFileName := fmt.Sprintf("doc_%d_v%d_%s", docID, newVersion, header.Filename)
// 	newFilePath := filepath.Join(uploadDir, newFileName)

// 	if err := os.WriteFile(newFilePath, fileBytes, 0644); err != nil {
// 		return fmt.Errorf("failed to save file: %w", err)
// 	}

// 	// 8. Insert new record into document_versions (version is stored as integer, but if your column is varchar, it will be auto-converted)
// 	queryInsert := `
// 		INSERT INTO document_versions (document_id, version, user_id, file_path, hashed_string)
// 		VALUES ($1, $2, $3, $4, $5)
// 	`
// 	_, err = tx.Exec(ctx, queryInsert, docID, newVersion, userID, newFilePath, newHash)
// 	if err != nil {
// 		return fmt.Errorf("failed to insert document version: %w", err)
// 	}

// 	// 9. Update the version column in the main documents table (FIXED – now uncommented)
// 	//    If the column is text, the integer will be accepted; if you want to be explicit, use $1::TEXT
// 	queryUpdateMain := `UPDATE documents SET version = $1 WHERE id = $2`
// 	_, err = tx.Exec(ctx, queryUpdateMain, newVersion, docID)
// 	if err != nil {
// 		return fmt.Errorf("failed to update main document version: %w", err)
// 	}

// 	// 10. Commit Transaction
// 	return tx.Commit(ctx)
// }
// func (s *Service) GetDocumentVersionsByDocumentID(ctx context.Context, db *pgxpool.Pool, documentID int64) ([]DocumentVersion, error) {
// 	versions, err := getDocumentVersionsByDocumentID(ctx, db, documentID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get document versions: %w", err)
// 	}
// 	return versions, nil
// }

package documentversion

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Repo    *Repository
	Storage *StorageManager
}

func NewService(repo *Repository, storage *StorageManager) *Service {
	return &Service{
		Repo:    repo,
		Storage: storage,
	}
}

/* ================================================================== */
/*  Create v1 — called once from document.Service.UploadDocument        */
/* ================================================================== */

// CreateDocumentVersion hashes + uploads the file to Cloudinary and inserts
// the v1 row. Returns the populated docVersion (ID + PublicID + Hash).
//
// NOTE: the caller must NOT have already uploaded the file — this method
// owns the Cloudinary upload for the version.
func (s *Service) CreateDocumentVersion(
	ctx context.Context,
	db *pgxpool.Pool,
	docVersion *DocumentVersion,
	file io.ReadSeeker,
) error {

	// 1. Hash
	hash, err := CalculateHash(file)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}
	docVersion.HashedString = hash

	// 2. Rewind so the upload sees the same bytes
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind: %w", err)
	}

	// 3. Upload to Cloudinary
	publicID, err := s.Storage.SaveVersionFile(
		docVersion.DocumentID,
		docVersion.Version,
		docVersion.OriginalFilename,
		file,
	)
	if err != nil {
		return fmt.Errorf("cloudinary upload: %w", err)
	}
	docVersion.PublicID = publicID

	// 4. Insert the version row (inside a tx)
	tx, err := db.Begin(ctx)
	if err != nil {
		s.Storage.DeleteFile(publicID)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 👇 original_filename is now saved too
	query := `
		INSERT INTO document_versions
			(user_id, document_id, version, public_id, original_filename, hashed_string)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at;
	`
	err = tx.QueryRow(ctx, query,
		docVersion.UserID,
		docVersion.DocumentID,
		docVersion.Version,
		docVersion.PublicID,
		docVersion.OriginalFilename,
		docVersion.HashedString,
	).Scan(&docVersion.ID, &docVersion.CreatedAt)
	if err != nil {
		s.Storage.DeleteFile(publicID)
		return fmt.Errorf("insert version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		s.Storage.DeleteFile(publicID)
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

/* ================================================================== */
/*  Update — creates v2, v3, ...                                       */
/* ================================================================== */

func (s *Service) UpdateDocumentVersion(
	ctx context.Context,
	db *pgxpool.Pool,
	docID int64,
	file multipart.File,
	header *multipart.FileHeader,
	userID int64,
) error {

	// 1. Read bytes (needed for hash + upload)
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	newHash, err := CalculateHash(bytes.NewReader(fileBytes))
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}

	// 2. Begin tx + lock the master row
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// version is now INTEGER — no ::INT cast needed
	var currentDocVersion int
	err = tx.QueryRow(ctx,
		`SELECT version FROM documents WHERE id = $1 FOR UPDATE`, docID,
	).Scan(&currentDocVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("document %d not found", docID)
		}
		return fmt.Errorf("lock document: %w", err)
	}

	// 3. Fetch latest version + hash
	var (
		currentVersion int
		currentHash    string
	)
	err = tx.QueryRow(ctx, `
		SELECT version, hashed_string
		FROM document_versions
		WHERE document_id = $1
		ORDER BY version DESC
		LIMIT 1
	`, docID).Scan(&currentVersion, &currentHash)

	hasRows := true
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			hasRows = false
		} else {
			return fmt.Errorf("fetch latest version: %w", err)
		}
	}

	// 4. Same hash → nothing changed
	if hasRows && currentHash == newHash {
		return nil
	}

	// 5. Compute next version number
	newVersion := 1
	if hasRows {
		newVersion = currentVersion + 1
	}

	// 6. Upload to Cloudinary (in memory)
	publicID, err := s.Storage.SaveVersionFile(
		docID,
		newVersion,
		header.Filename,
		bytes.NewReader(fileBytes),
	)
	if err != nil {
		return fmt.Errorf("cloudinary upload: %w", err)
	}

	// 7. Insert version row — includes original_filename
	_, err = tx.Exec(ctx, `
		INSERT INTO document_versions
			(document_id, version, user_id, public_id, original_filename, hashed_string)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, docID, newVersion, userID, publicID, header.Filename, newHash)
	if err != nil {
		s.Storage.DeleteFile(publicID)
		return fmt.Errorf("insert version: %w", err)
	}

	// 8. Re-point the master row at the new version
	_, err = tx.Exec(ctx,
		`UPDATE documents SET version = $1, public_id = $2 WHERE id = $3`,
		newVersion, publicID, docID,
	)

	// 9. Commit
	if err := tx.Commit(ctx); err != nil {
		s.Storage.DeleteFile(publicID)
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

/* ================================================================== */
/*  Read                                                               */
/* ================================================================== */

func (s *Service) GetDocumentVersionsByDocumentID(
	ctx context.Context,
	db *pgxpool.Pool,
	documentID int64,
) ([]DocumentVersion, error) {
	return getDocumentVersionsByDocumentID(ctx, db, documentID)
}

func (s *Service) GetVersionByID(
	ctx context.Context,
	versionID string,
) (*DocumentVersion, error) {
	return getDocumentVersionByID(ctx, s.Repo.DB, versionID)
}

func (s *Service) GetAllVersions(
	ctx context.Context,
) ([]DocumentVersion, error) {
	return getAllDocumentVersions(ctx, s.Repo.DB)
}

/* ================================================================== */
/*  Restore                                                            */
/* ================================================================== */

// RestoreVersion re-points the master document to an older version.
// Does NOT create a new row — the old version already exists.
func (s *Service) RestoreVersion(
	ctx context.Context,
	docID int64,
	version int,
	publicID string,
) error {
	tag, err := s.Repo.DB.Exec(ctx, `
		UPDATE documents
		SET version = $1, file_path = $2
		WHERE id = $3
	`, version, publicID, docID)
	if err != nil {
		return fmt.Errorf("restore version: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("document not found")
	}
	return nil
}

/* ================================================================== */
/*  Delete                                                             */
/* ================================================================== */

// DeleteVersion removes one version. Refuses to delete the current version.
func (s *Service) DeleteVersion(
	ctx context.Context,
	versionID string,
) error {
	v, err := getDocumentVersionByID(ctx, s.Repo.DB, versionID)
	if err != nil {
		return fmt.Errorf("fetch version: %w", err)
	}

	// Refuse if it's the current version
	var currentVersion int
	err = s.Repo.DB.QueryRow(ctx,
		`SELECT version FROM documents WHERE id = $1`, v.DocumentID,
	).Scan(&currentVersion)
	if err == nil && v.Version == currentVersion {
		return errors.New("cannot delete the current version")
	}

	// Delete DB row
	if _, err := s.Repo.DB.Exec(ctx,
		`DELETE FROM document_versions WHERE id = $1`, versionID,
	); err != nil {
		return fmt.Errorf("delete version row: %w", err)
	}

	// Best-effort delete from Cloudinary
	s.Storage.DeleteFile(v.PublicID)

	return nil
}
