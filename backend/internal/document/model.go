package document

import "time"

type Document struct {
	ID          int64     `json:"id"`
	FolderID    int64     `json:"folder_id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	FilePath    string    `json:"file_path"`
	MimeType    string    `json:"mime_type"`
	Size        int64     `json:"size"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Description string    `json:"description,omitempty"`
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	Comment     string    `json:"comment,omitempty"`
}
