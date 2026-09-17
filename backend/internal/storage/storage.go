// package storage

// // internal/storage/storage.go

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"time"
// )

// func SaveFile(userID int64, filename string, content io.Reader) (string, error) {
// 	now := time.Now()
// 	dir := filepath.Join("uploads", now.Format("2006"), now.Format("01"), now.Format("02"))
// 	if err := os.MkdirAll(dir, 0755); err != nil {
// 		return "", err
// 	}
// 	uniqueName := fmt.Sprintf("%d_%d_%s", userID, now.UnixNano(), filename)
// 	filePath := filepath.Join(dir, uniqueName)
// 	out, err := os.Create(filePath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer out.Close()
// 	_, err = io.Copy(out, content)
// 	return filePath, err
// }

// func DeleteFile(filePath string) error {
// 	return os.Remove(filePath)
// }

package storage

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

/* ------------------------------------------------------------------ */
/*  Save                                                              */
/* ------------------------------------------------------------------ */

func SaveFile(userID int64, filename string, content io.Reader) (string, error) {
	cld, err := cloudinary.New()
	if err != nil {
		return "", fmt.Errorf("cloudinary init: %w", err)
	}
	cld.Config.URL.Secure = true

	ext := filepath.Ext(filename)                       // e.g. ".pdf"
	base := sanitize(strings.TrimSuffix(filename, ext)) // e.g. "abdulfetah-cv"
	resourceType := getResourceType(filename)           // "image" for PDFs & images, "raw" for docs/zips

	// Build clean publicID maintaining file extension
	publicID := fmt.Sprintf("user_%d/%s_%d%s", userID, base, time.Now().UnixNano(), ext)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	resp, err := cld.Upload.Upload(ctx, content, uploader.UploadParams{
		Folder:       "uploads",
		PublicID:     publicID,
		ResourceType: resourceType,
		Overwrite:    boolPtr(true),
		Invalidate:   boolPtr(true),
	})
	fmt.Println("Cloudinary upload response:", resp.SecureURL)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload: %w", err)
	}

	// Returns e.g. "uploads/user_1/abdulfetah-cv_1789655505762953444.pdf"
	return resp.SecureURL, nil
}

/* ------------------------------------------------------------------ */
/*  Get URL                                                           */
/* ------------------------------------------------------------------ */

// GetFileURL builds a Cloudinary delivery URL based on publicID, resource type, and download preference.
func GetFileURL(publicID string, isRaw bool, forceDownload bool) (string, error) {
	if publicID == "" {
		return "", fmt.Errorf("public_id is empty")
	}

	// cld, err := cloudinary.New()
	// if err != nil {
	// 	return "", fmt.Errorf("cloudinary init: %w", err)
	// }
	// cld.Config.URL.Secure = true

	// Clean up legacy database entries containing double extensions (.pdf.pdf)
	// cleanPublicID := publicID
	// for strings.HasSuffix(cleanPublicID, ".pdf.pdf") {
	// 	cleanPublicID = strings.TrimSuffix(cleanPublicID, ".pdf")
	// }

	// True raw files (.docx, .xlsx, .zip, .rar, .txt)
	// if isRaw {
	// 	asset, err := cld.File(cleanPublicID)
	// 	if err != nil {
	// 		return "", fmt.Errorf("cloudinary raw asset: %w", err)
	// 	}
	// 	return asset.String()
	// }

	// Images, Videos, and PDFs (served via Cloudinary Image pipeline)
	// asset, err := cld.Image(cleanPublicID)
	// if err != nil {
	// 	return "", fmt.Errorf("cloudinary image asset: %w", err)
	// }

	// Attachment flag forces browser to download the file instead of inline preview
	// if forceDownload {
	// 	asset.Transformation = "fl_attachment"
	// }

	return publicID, nil
}

/* ------------------------------------------------------------------ */
/*  Delete                                                            */
/* ------------------------------------------------------------------ */

func DeleteFile(publicID string) error {
	if publicID == "" {
		return nil
	}

	cld, err := cloudinary.New()
	if err != nil {
		return fmt.Errorf("cloudinary init: %w", err)
	}

	// cleanPublicID := publicID
	// for strings.HasSuffix(cleanPublicID, ".pdf.pdf") {
	// 	cleanPublicID = strings.TrimSuffix(cleanPublicID, ".pdf")
	// }

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: getResourceType(publicID),
		Invalidate:   boolPtr(true),
	})
	if err != nil {
		return fmt.Errorf("cloudinary delete: %w", err)
	}
	return nil
}

/* ------------------------------------------------------------------ */
/*  Helpers                                                           */
/* ------------------------------------------------------------------ */

func getResourceType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	// Cloudinary treats PDFs as "image" assets for page rendering, previews, and attachments
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".tiff", ".ico", ".pdf":
		return "image"
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".mp3", ".wav":
		return "video"
	case ".docx", ".doc", ".xlsx", ".xls", ".pptx", ".zip", ".rar", ".txt", ".csv":
		return "raw"
	default:
		if ext == "" {
			return "image"
		}
		return "raw"
	}
}

func sanitize(name string) string {
	replacer := strings.NewReplacer(
		" ", "_", "/", "_", "\\", "_",
		"#", "_", "?", "_", "&", "_",
	)
	return replacer.Replace(name)
}

func boolPtr(b bool) *bool { return &b }
