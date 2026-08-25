package folder

import "time"

type Folder struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateFolderRequest struct {
	Name string `json:"name"`
}
