package storage

// internal/storage/storage.go

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func SaveFile(userID int64, filename string, content io.Reader) (string, error) {
	now := time.Now()
	dir := filepath.Join("uploads", now.Format("2006"), now.Format("01"), now.Format("02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	uniqueName := fmt.Sprintf("%d_%d_%s", userID, now.UnixNano(), filename)
	filePath := filepath.Join(dir, uniqueName)
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, content)
	return filePath, err
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}
