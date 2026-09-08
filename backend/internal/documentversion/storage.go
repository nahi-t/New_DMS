package documentversion

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type StorageManager struct {
	baseDir string
}

func NewStorageManager(baseDir string) *StorageManager {
	return &StorageManager{baseDir: baseDir}
}

// SaveVersionFile strictly handles file system I/O.
func (sm *StorageManager) SaveVersionFile(
	docID int64,
	version int,
	originalFilename string,
	src io.ReadSeeker,
) (string, error) {

	docDir := filepath.Join(sm.baseDir, "documents", fmt.Sprintf("%d", docID))
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create document directory: %w", err)
	}

	uniqueID := time.Now().UnixNano()
	uniqueFilename := fmt.Sprintf("v%d_%d_%s", version, uniqueID, originalFilename)
	filePath := filepath.Join(docDir, uniqueFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer dst.Close()

	// Pure disk write
	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(filePath)
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	return filePath, nil
}

func (sm *StorageManager) DeleteFile(filePath string) {
	if filePath != "" {
		_ = os.Remove(filePath)
	}
}
