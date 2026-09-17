package document

import "time"

type Document struct {
	ID       int64  `json:"id"         db:"id"`
	FolderID int64  `json:"folder_id"  db:"folder_id"`
	UserID   int64  `json:"user_id"    db:"user_id"`
	Name     string `json:"name"       db:"name"`

	// Cloudinary public_id — stored in the `file_path` DB column
	// (kept the old column name to avoid a migration).
	PublicID string `json:"public_id" db:"file_path"`

	MimeType    string    `json:"mime_type"      db:"mime_type"`
	Size        int64     `json:"size"           db:"size"`
	UploadedAt  time.Time `json:"uploaded_at"    db:"uploaded_at"`
	Description string    `json:"description,omitempty" db:"description"`
	Version     int       `json:"version"        db:"version"`
	Status      string    `json:"status"         db:"status"`
	Comment     string    `json:"comment"        db:"comment"`
}
