// package documentversion

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"time"
// )

// type StorageManager struct {
// 	baseDir string
// }

// func NewStorageManager(baseDir string) *StorageManager {
// 	return &StorageManager{baseDir: baseDir}
// }

// // SaveVersionFile strictly handles file system I/O.
// func (sm *StorageManager) SaveVersionFile(
// 	docID int64,
// 	version int,
// 	originalFilename string,
// 	src io.ReadSeeker,
// ) (string, error) {

// 	docDir := filepath.Join(sm.baseDir, "documents", fmt.Sprintf("%d", docID))
// 	if err := os.MkdirAll(docDir, 0755); err != nil {
// 		return "", fmt.Errorf("failed to create document directory: %w", err)
// 	}

// 	uniqueID := time.Now().UnixNano()
// 	uniqueFilename := fmt.Sprintf("v%d_%d_%s", version, uniqueID, originalFilename)
// 	filePath := filepath.Join(docDir, uniqueFilename)

// 	dst, err := os.Create(filePath)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create file on disk: %w", err)
// 	}
// 	defer dst.Close()

// 	// Pure disk write
// 	if _, err := io.Copy(dst, src); err != nil {
// 		_ = os.Remove(filePath)
// 		return "", fmt.Errorf("failed to save file content: %w", err)
// 	}

// 	return filePath, nil
// }

// func (sm *StorageManager) DeleteFile(filePath string) {
// 	if filePath != "" {
// 		_ = os.Remove(filePath)
// 	}
// }

package documentversion

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type StorageManager struct {
	cld    *cloudinary.Cloudinary
	folder string // base folder inside Cloudinary (e.g. "documents")
}

// NewStorageManager initializes Cloudinary from the CLOUDINARY_URL env var.
// folder = base folder in Cloudinary, e.g. "documents".
func NewStorageManager(folder string) (*StorageManager, error) {
	cld, err := cloudinary.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init cloudinary: %w", err)
	}
	cld.Config.URL.Secure = true

	return &StorageManager{
		cld:    cld,
		folder: folder,
	}, nil
}

// SaveVersionFile uploads a document version to Cloudinary and returns its
// public_id. Store this public_id in your DB.
//
// IMPORTANT: For raw files, the extension MUST be part of the public_id,
// otherwise the resulting URL won't have it and browsers won't download it.
func (sm *StorageManager) SaveVersionFile(
	docID int64,
	version int,
	originalFilename string,
	src io.ReadSeeker,
) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Build a unique public_id that keeps the original extension.
	ext := filepath.Ext(originalFilename)             // ".pdf", ".docx", etc.
	base := strings.TrimSuffix(originalFilename, ext) // "report"
	base = sanitize(base)                             // strip spaces / slashes
	publicID := fmt.Sprintf("v%d_%d_%s%s", version, time.Now().UnixNano(), base, ext)

	resp, err := sm.cld.Upload.Upload(ctx, src, uploader.UploadParams{
		Folder:         fmt.Sprintf("%s/%d", sm.folder, docID), // documents/<docID>
		PublicID:       publicID,
		ResourceType:   "raw", // documents = raw
		UseFilename:    boolPtr(false),
		UniqueFilename: boolPtr(false), // we already made it unique
		Overwrite:      boolPtr(false),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to cloudinary: %w", err)
	}

	// Return the public_id — this is what you store in DB and use to delete later.
	return resp.SecureURL, nil
}

// DeleteFile removes the asset from Cloudinary by its public_id.
// MUST use ResourceType "raw" to match the upload.
func (sm *StorageManager) DeleteFile(publicID string) {
	if publicID == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := sm.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "raw", // must match upload
		Invalidate:   boolPtr(true),
	})
	if err != nil {
		// Log but don't panic — deletion failures shouldn't crash the caller.
		fmt.Printf("cloudinary delete failed for %s: %v\n", publicID, err)
	}
}

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

// sanitize makes a filename safe to embed in a Cloudinary public_id.
func sanitize(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "#", "_")
	name = strings.ReplaceAll(name, "?", "_")
	return name
}

func boolPtr(b bool) *bool { return &b }
