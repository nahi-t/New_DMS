package documentversion

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// CalculateHash computes the SHA-256 hex string of any io.Reader stream.
func CalculateHash(r io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// SaveAndHashFile streams an uploaded file to the destination path while
// simultaneously computing its SHA-256 hash in a single memory-efficient pass.
func SaveAndHashFile(src io.Reader, destPath string) (string, error) {
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	hasher := sha256.New()
	// MultiWriter streams incoming bytes to both the target file and the hasher at the same time
	writer := io.MultiWriter(dst, hasher)

	if _, err := io.Copy(writer, src); err != nil {
		_ = os.Remove(destPath) // Clean up partial file on error
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
