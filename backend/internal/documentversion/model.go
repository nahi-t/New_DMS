package documentversion

import "time"

type DocumentVersion struct {
	ID           string    `json:"id" db:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       int64     `json:"user_id" db:"user_id" gorm:"type:bigint;not null;index"`
	DocumentID   int64     `json:"document_id" db:"document_id" gorm:"type:bigint;not null;index"`
	Version      int       `json:"version" db:"version" gorm:"type:varchar(20);not null"`
	FilePath     string    `json:"file_path" db:"file_path" gorm:"type:text;not null"`
	Time         time.Time `json:"time" db:"time" gorm:"autoCreateTime"`
	HashedString string    `json:"hashed_string" db:"hashed_string" gorm:"type:text;not null"`
}
