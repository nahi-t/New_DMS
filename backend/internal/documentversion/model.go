package documentversion

import "time"

type DocumentVersion struct {
	ID               string    `json:"id"                db:"id"                gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID           int64     `json:"user_id"           db:"user_id"           gorm:"type:bigint;not null;index"`
	DocumentID       int64     `json:"document_id"       db:"document_id"       gorm:"type:bigint;not null;index"`
	Version          int       `json:"version"           db:"version"           gorm:"type:integer;not null"` // 👈 integer, not varchar
	PublicID         string    `json:"public_id"         db:"public_id"         gorm:"type:text;not null"`    // 👈 was FilePath
	OriginalFilename string    `json:"original_filename" db:"original_filename" gorm:"type:text"`             // 👈 new
	HashedString     string    `json:"hashed_string"     db:"hashed_string"     gorm:"type:text;not null"`
	CreatedAt        time.Time `json:"created_at"        db:"created_at"        gorm:"autoCreateTime"` // 👈 was Time
}
