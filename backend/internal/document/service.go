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
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service holds dependencies for document business logic.
type Service struct {
	DB *pgxpool.Pool
}

// NewService creates a new document service.
func NewService(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

// UploadDocument handles the complete upload process.
func (s *Service) UploadDocument(
	ctx context.Context,
	folderID int64,
	uploader *user.User,
	fileHeader *multipart.FileHeader,
	file multipart.File,
	description string,
) (*Document, error) {
	// if uploader.Role != user.RoleAdmin && uploader.Role != user.RoleManager {
	// 	return nil, errors.New("permission denied: only admins and managers can upload")
	// }

	filePath, err := storage.SaveFile(uploader.ID, fileHeader.Filename, file)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		FolderID:    folderID,
		UserID:      uploader.ID,
		Name:        fileHeader.Filename,
		FilePath:    filePath,
		MimeType:    fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
		Description: description,
		Version:     1, // Initial version
	}

	// 1. Insert master document to get the document ID (doc.ID)
	if err := InsertDocument(ctx, s.DB, doc); err != nil {
		storage.DeleteFile(filePath)
		return nil, err
	}

	// 2. Rewind the file stream back to start before versioning/hashing
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		storage.DeleteFile(filePath)
		return nil, err
	}

	// 3. Build the DocumentVersion struct
	docVersion := &documentversion.DocumentVersion{
		UserID:     uploader.ID,
		DocumentID: doc.ID, // 👈 Using int64 directly
		Version:    1,
		FilePath:   filePath,
	}

	// 4. Call the package-level CreateDocumentVersion function cleanly
	if err := documentversion.CreateDocumentVersion(ctx, s.DB, docVersion, file); err != nil {
		storage.DeleteFile(filePath)
		return nil, err
	}

	return doc, nil
}

// ListDocuments returns documents for a folder.
func (s *Service) ListDocuments(ctx context.Context, folderID int64, viewer *user.User) ([]Document, error) {
	return GetDocumentsByFolder(ctx, s.DB, folderID)
}

// DownloadDocument retrieves and authorizes download.
func (s *Service) DownloadDocument(ctx context.Context, docID int64, requester *user.User) (*Document, error) {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return nil, err
	}
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin && requester.Role != user.RoleManager {
		return nil, errors.New("permission denied")
	}
	return doc, nil
}

// DeleteDocument removes document (DB + file) with permission check.
func (s *Service) DeleteDocument(ctx context.Context, docID int64, requester *user.User) error {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}
	if err := DeleteDocumentRecord(ctx, s.DB, docID); err != nil {
		return err
	}
	storage.DeleteFile(doc.FilePath)
	return nil
}

// RenameDocument changes the name of a document.
func (s *Service) RenameDocument(ctx context.Context, docID int64, newName string, requester *user.User) error {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}
	return UpdateDocumentName(ctx, s.DB, docID, newName)
}

func (s *Service) updateDocumentContent(ctx context.Context, docID int64, newFile multipart.File, newFileHeader *multipart.FileHeader, userID int64, userRole string) error {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}

	// Check permissions using the role and ID from context
	if userID != doc.UserID && userRole != "admin" { // Adjust "admin" to match your role string constant
		return errors.New("permission denied")
	}

	return documentversion.UpdateDocumentVersion(ctx, s.DB, docID, newFile, newFileHeader, userID)
}

// MoveDocument changes the folder of a document.
func (s *Service) MoveDocument(ctx context.Context, docID, newFolderID int64, requester *user.User) error {
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}
	if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
		return errors.New("permission denied")
	}
	return MoveDocument(ctx, s.DB, docID, newFolderID)
}

// SearchDocuments returns search results.
func (s *Service) SearchDocuments(ctx context.Context, searchTerm string, folderID *int64, requester *user.User) ([]Document, error) {
	return SearchDocuments(ctx, s.DB, searchTerm, folderID)
}

func (s *Service) UpdateStatus(
	ctx context.Context,
	docID int64,
	newStatus string,
	requester *user.User,
	comment string,
) error {

	// 1. Get authenticated user ID from context
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 2. Get the document
	doc, err := GetDocumentByID(ctx, s.DB, docID)
	if err != nil {
		return err
	}

	if doc == nil {
		return errors.New("document not found")
	}

	// 3. Make sure requester exists
	if requester == nil {
		return errors.New("requester is nil")
	}

	// 4. Make sure the authenticated user matches requester
	if requester.ID != userID {
		return errors.New("authenticated user mismatch")
	}

	// 5. Check permission
	// if requester.ID != doc.UserID && requester.Role != user.RoleAdmin {
	// 	return errors.New("permission denied")
	// }

	// 6. Update document status
	folderID, err := UpdateDocumentStatus(
		ctx,
		s.DB,
		docID,
		newStatus,
		comment,
	)
	if err != nil {
		return err
	}

	// 7. Get the user associated with the folder
	folderUserID, err := FatchUserFromFOlder(
		ctx,
		s.DB,
		folderID,
	)
	if err != nil {
		return err
	}

	// 8. Check folder permission
	if folderUserID != userID {
		return errors.New("permission denied")
	}

	return nil
}
