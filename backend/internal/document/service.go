// package document

// import (
// 	"context"
// 	"errors"

// 	"io"
// 	"mime/multipart"

// 	"github.com/docmanage_new/internal/auth"
// 	"github.com/docmanage_new/internal/documentversion"
// 	"github.com/docmanage_new/internal/storage"
// 	"github.com/docmanage_new/internal/user"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// // Service holds dependencies for document business logic.
// type Service struct {
// 	DB *pgxpool.Pool
// }

// // NewService creates a new document service.
// func NewService(db *pgxpool.Pool) *Service {
// 	return &Service{DB: db}
// }

// // UploadDocument handles the complete upload process.
// func (s *Service) UploadDocument(
// 	ctx context.Context,
// 	folderID int64,
// 	uploader *user.User,
// 	fileHeader *multipart.FileHeader,
// 	file multipart.File,
// 	description string,
// ) (*Document, error) {
// 	// if uploader.Role != user.RoleAdmin && uploader.Role != user.RoleManager {
// 	// 	return nil, errors.New("permission denied: only admins and managers can upload")
// 	// }

// 	filePath, err := storage.SaveFile(uploader.ID, fileHeader.Filename, file)
// 	if err != nil {
// 		return nil, err
// 	}

// 	doc := &Document{
// 		FolderID:    folderID,
// 		UserID:      uploader.ID,
// 		Name:        fileHeader.Filename,
// 		FilePath:    filePath,
// 		MimeType:    fileHeader.Header.Get("Content-Type"),
// 		Size:        fileHeader.Size,
// 		Description: description,
// 		Version:     1, // Initial version
// 	}

// 	// 1. Insert master document to get the document ID (doc.ID)
// 	if err := InsertDocument(ctx, s.DB, doc); err != nil {
// 		storage.DeleteFile(filePath)
// 		return nil, err
// 	}

// 	// 2. Rewind the file stream back to start before versioning/hashing
// 	if _, err := file.Seek(0, io.SeekStart); err != nil {
// 		storage.DeleteFile(filePath)
// 		return nil, err
// 	}

// 	// 3. Build the DocumentVersion struct
// 	docVersion := &documentversion.DocumentVersion{
// 		UserID:     uploader.ID,
// 		DocumentID: doc.ID, // 👈 Using int64 directly
// 		Version:    1,
// 		FilePath:   public,
// 	}

// 	// 4. Call the package-level CreateDocumentVersion function cleanly
// 	if err := documentversion.CreateDocumentVersion(ctx, s.DB, docVersion, file); err != nil {
// 		storage.DeleteFile(filePath)
// 		return nil, err
// 	}

// 	return doc, nil
// }

// // ListDocuments returns documents for a folder.
// func (s *Service) ListDocuments(ctx context.Context, folderID int64, viewer *user.User) ([]Document, error) {
// 	return GetDocumentsByFolder(ctx, s.DB, folderID)
// }

// // DownloadDocument retrieves and authorizes download.
// func (s *Service) DownloadDocument(ctx context.Context, docID int64, requester *user.User) (*Document, error) {
// 	doc, err := GetDocumentByID(ctx, s.DB, docID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin && requester.Role != user.RoleManager {
// 		return nil, errors.New("permission denied")
// 	}
// 	return doc, nil
// }

// // DeleteDocument removes document (DB + file) with permission check.
// func (s *Service) DeleteDocument(ctx context.Context, docID int64, requester *user.User) error {
// 	doc, err := GetDocumentByID(ctx, s.DB, docID)
// 	if err != nil {
// 		return err
// 	}
// 	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
// 		return errors.New("permission denied")
// 	}
// 	if err := DeleteDocumentRecord(ctx, s.DB, docID); err != nil {
// 		return err
// 	}
// 	storage.DeleteFile(doc.FilePath)
// 	return nil
// }

// // RenameDocument changes the name of a document.
// func (s *Service) RenameDocument(ctx context.Context, docID int64, newName string, requester *user.User) error {
// 	doc, err := GetDocumentByID(ctx, s.DB, docID)
// 	if err != nil {
// 		return err
// 	}
// 	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
// 		return errors.New("permission denied")
// 	}
// 	return UpdateDocumentName(ctx, s.DB, docID, newName)
// }

// func (s *Service) updateDocumentContent(ctx context.Context, docID int64, newFile multipart.File, newFileHeader *multipart.FileHeader, userID int64, userRole string) error {
// 	doc, err := GetDocumentByID(ctx, s.DB, docID)
// 	if err != nil {
// 		return err
// 	}

// 	// Check permissions using the role and ID from context
// 	if userID != doc.UserID && userRole != "admin" { // Adjust "admin" to match your role string constant
// 		return errors.New("permission denied")
// 	}

// 	return documentversion.UpdateDocumentVersion(ctx, s.DB, docID, newFile, newFileHeader, userID)
// }

// // MoveDocument changes the folder of a document.
// func (s *Service) MoveDocument(ctx context.Context, docID, newFolderID int64, requester *user.User) error {
// 	doc, err := GetDocumentByID(ctx, s.DB, docID)
// 	if err != nil {
// 		return err
// 	}
// 	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
// 		return errors.New("permission denied")
// 	}
// 	return MoveDocument(ctx, s.DB, docID, newFolderID)
// }

// // SearchDocuments returns search results.
// func (s *Service) SearchDocuments(ctx context.Context, searchTerm string, folderID *int64, requester *user.User) ([]Document, error) {
// 	return SearchDocuments(ctx, s.DB, searchTerm, folderID)
// }

// func (s *Service) UpdateStatus(
// 	ctx context.Context,
// 	docID int64,
// 	newStatus string,
// 	requester *user.User,
// 	comment string,
// ) error {

// 	// 1. Get authenticated user ID from context
// 	userID, err := auth.GetUserIDFromContext(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// 2. Get the document
// 	doc, err := GetDocumentByID(ctx, s.DB, docID)
// 	if err != nil {
// 		return err
// 	}

// 	if doc == nil {
// 		return errors.New("document not found")
// 	}

// 	// 3. Make sure requester exists
// 	if requester == nil {
// 		return errors.New("requester is nil")
// 	}

// 	// 4. Make sure the authenticated user matches requester
// 	if requester.ID != userID {
// 		return errors.New("authenticated user mismatch")
// 	}

// 	// 5. Check permission
// 	// if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
// 	// 	return errors.New("permission denied")
// 	// }

// 	// 6. Update document status
// 	folderID, err := UpdateDocumentStatus(
// 		ctx,
// 		s.DB,
// 		docID,
// 		newStatus,
// 		comment,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	// 7. Get the user associated with the folder
// 	folderUserID, err := FatchUserFromFOlder(
// 		ctx,
// 		s.DB,
// 		folderID,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	// 8. Check folder permission
// 	if folderUserID != userID {
// 		return errors.New("permission denied")
// 	}

// 	return nil
// }

package document

import (
	"context"
	"errors"
	"io"
	"mime/multipart"

	"github.com/docmanage_new/internal/auth"
	"github.com/docmanage_new/internal/documentversion"
	"github.com/docmanage_new/internal/storage"
	"github.com/docmanage_new/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ================================================================== */
/*  Service                                                            */
/* ================================================================== */

// Service holds dependencies for document business logic.
type Service struct {
	DB       *pgxpool.Pool
	Versions *documentversion.Service // handles version rows
}

// NewService wires up the document service.
func NewService(db *pgxpool.Pool, versions *documentversion.Service) *Service {
	return &Service{
		DB:       db,
		Versions: versions,
	}
}

/* ================================================================== */
/*  Upload — first version                                             */
/* ================================================================== */

// UploadDocument handles the complete first-time upload.
//
// Flow:
//  1. Hash the file (for dedup on the next update)
//  2. Rewind and upload to Cloudinary (ONE upload only)
//  3. Insert `documents` row (current pointer)
//  4. Insert `document_versions` row (v1)
//  5. Commit — both rows succeed or neither does
func (s *Service) UploadDocument(
	ctx context.Context,
	folderID int64,
	uploader *user.User,
	fileHeader *multipart.FileHeader,
	file multipart.File,
	description string,
) (*Document, error) {

	// 1. Hash the file BEFORE we consume the stream.
	hash, err := documentversion.CalculateHash(file)
	if err != nil {
		return nil, err
	}

	// 2. Rewind so the upload sees the same bytes.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// 3. Upload to Cloudinary — this is the ONLY upload on first save.
	publicID, err := storage.SaveFile(uploader.ID, fileHeader.Filename, file)
	if err != nil {
		return nil, err
	}

	// 4. Begin a transaction so both inserts are atomic.
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		storage.DeleteFile(publicID)
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 5. Insert the master document row.
	doc := &Document{
		FolderID:    folderID,
		UserID:      uploader.ID,
		Name:        fileHeader.Filename,
		PublicID:    publicID,
		MimeType:    fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Description: description,
		Version:     1,
	}
	if err := InsertDocumentTx(ctx, tx, doc); err != nil {
		storage.DeleteFile(publicID)
		return nil, err
	}

	// 6. Insert the v1 row in document_versions.
	version := &documentversion.DocumentVersion{
		UserID:           uploader.ID,
		DocumentID:       doc.ID,
		Version:          1,
		PublicID:         publicID, // same file as the master
		OriginalFilename: fileHeader.Filename,
		HashedString:     hash,
	}
	if err := s.Versions.Repo.InsertDocumentVersion(ctx, tx, version, hash); err != nil {
		storage.DeleteFile(publicID)
		return nil, err
	}

	// 7. Commit.
	if err := tx.Commit(ctx); err != nil {
		storage.DeleteFile(publicID)
		return nil, err
	}

	return doc, nil
}

/* ================================================================== */
/*  Read                                                               */
/* ================================================================== */

// ListDocuments returns documents for a folder.
func (s *Service) ListDocuments(
	ctx context.Context,
	folderID int64,
	viewer *user.User,
) ([]Document, error) {
	return GetDocumentsByFolder(ctx, s.DB, folderID)
}

// GetDocument returns a single document by ID.
func (s *Service) GetDocument(
	ctx context.Context,
	docID int64,
	requester *user.User,
) (*Document, error) {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("document not found")
	}
	return doc, nil
}

// DownloadDocument retrieves a document after checking permissions.
// Returns the document metadata; the caller builds the Cloudinary URL.
func (s *Service) DownloadDocument(
	ctx context.Context,
	docID int64,
	requester *user.User,
) (*Document, error) {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("document not found")
	}
	if requester.ID != doc.UserID &&
		requester.Role != user.RoleAdmin &&
		requester.Role != user.RoleManager {
		return nil, errors.New("permission denied")
	}
	return doc, nil
}

// SearchDocuments returns search results within a folder (or globally if folderID is nil).
func (s *Service) SearchDocuments(
	ctx context.Context,
	searchTerm string,
	folderID *int64,
	requester *user.User,
) ([]Document, error) {
	return SearchDocuments(ctx, s.DB, searchTerm, folderID)
}

/* ================================================================== */
/*  Update — new version                                               */
/* ================================================================== */

// UpdateDocumentContent uploads a new version of the file and updates the
// master `documents` row to point at it.
//
// The full flow (hash → dedup → Cloudinary upload → insert version row →
// update master row) lives inside documentversion.Service.UpdateDocumentVersion.
func (s *Service) UpdateDocumentContent(
	ctx context.Context,
	docID int64,
	newFile multipart.File,
	newFileHeader *multipart.FileHeader,
	requester *user.User,
) error {

	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}

	// Permission check
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}

	// Delegate the entire update flow.
	return s.Versions.UpdateDocumentVersion(
		ctx,
		s.DB,
		docID,
		newFile,
		newFileHeader,
		requester.ID,
	)
}

// RenameDocument changes the display name of a document.
func (s *Service) RenameDocument(
	ctx context.Context,
	docID int64,
	newName string,
	requester *user.User,
) error {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}
	return UpdateDocumentName(ctx, s.DB, docID, newName)
}

// MoveDocument changes the folder of a document.
func (s *Service) MoveDocument(
	ctx context.Context,
	docID, newFolderID int64,
	requester *user.User,
) error {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}
	return MoveDocument(ctx, s.DB, docID, newFolderID)
}

/* ================================================================== */
/*  Status workflow                                                    */
/* ================================================================== */

// UpdateStatus advances the workflow status of a document.
func (s *Service) UpdateStatus(
	ctx context.Context,
	docID int64,
	newStatus string,
	requester *user.User,
	comment string,
) error {

	// 1. Authenticated user from context
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 2. Fetch the document
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}

	// 3. Sanity-check requester
	if requester == nil {
		return errors.New("requester is nil")
	}
	if requester.ID != userID {
		return errors.New("authenticated user mismatch")
	}

	// 4. Update status — also returns the folder_id for the permission check
	folderID, err := UpdateDocumentStatus(ctx, s.DB, docID, newStatus, comment)
	if err != nil {
		return err
	}

	// 5. Admin bypasses the folder-owner check entirely
	if requester.Role == user.RoleAdmin {
		return nil
	}

	// 6. Managers: allow if they own the folder OR the document
	if requester.Role == user.RoleManager {
		if doc.UserID == requester.ID {
			return nil
		}
		folderUserID, err := FatchUserFromFOlder(ctx, s.DB, folderID)
		if err != nil {
			return err
		}
		if folderUserID == userID {
			return nil
		}
		return errors.New("permission denied: not your folder")
	}

	// 7. Regular users: must own the folder or the document
	if doc.UserID == requester.ID {
		return nil
	}
	folderUserID, err := FatchUserFromFOlder(ctx, s.DB, folderID)
	if err != nil {
		return err
	}
	if folderUserID != userID {
		return errors.New("permission denied: not your folder")
	}

	return nil
}

/* ================================================================== */
/*  Delete                                                             */
/* ================================================================== */

// DeleteDocument removes a document from the DB and deletes its Cloudinary
// assets (both the master and every version).
func (s *Service) DeleteDocument(
	ctx context.Context,
	docID int64,
	requester *user.User,
) error {

	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}

	// Permission: owner or admin only
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}

	// 1. Collect every Cloudinary public_id BEFORE deleting rows.
	versions, err := s.Versions.GetDocumentVersionsByDocumentID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	publicIDs := make([]string, 0, len(versions)+1)
	for _, v := range versions {
		if v.PublicID != "" {
			publicIDs = append(publicIDs, v.PublicID)
		}
	}
	// Also include the master row's pointer, in case it's different.
	if doc.PublicID != "" {
		publicIDs = append(publicIDs, doc.PublicID)
	}

	// 2. Delete the DB rows first — a tx keeps them consistent.
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM document_versions WHERE document_id = $1`, docID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM documents WHERE id = $1`, docID,
	); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// 3. Best-effort delete from Cloudinary. If this fails, we still
	//    removed the DB rows — the asset just stays orphaned.
	for _, id := range dedupe(publicIDs) {
		_ = storage.DeleteFile(id)
	}

	return nil
}

/* ================================================================== */
/*  Helpers                                                            */
/* ================================================================== */

// dedupe removes duplicate strings (e.g. when master + v1 share a public_id).
func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := in[:0]
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// ensure pgx is used if you later add tx-typed helpers here.
var _ = pgx.ErrNoRows
