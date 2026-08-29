package document

import (
	"context"
	"errors"
	"mime/multipart"

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
	if uploader.Role != user.RoleAdmin && uploader.Role != user.RoleManager {
		return nil, errors.New("permission denied: only admins and managers can upload")
	}

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
		Version:     1,
	}

	if err := InsertDocument(ctx, s.DB, doc); err != nil {
		storage.DeleteFile(filePath)
		return nil, err
	}
	return doc, nil
}

// ListDocuments returns documents for a folder.
func (s *Service) ListDocuments(ctx context.Context, folderID int64, viewer *user.User) ([]Document, error) {
	// Optional: check folder visibility for the viewer
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
