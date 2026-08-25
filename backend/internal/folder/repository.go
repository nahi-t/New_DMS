package folder

import (
	"database/sql"
	"time"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(f *Folder) error {
	query := `INSERT INTO folders (name, created_by, created_at) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.DB.QueryRow(query, f.Name, f.CreatedBy, time.Now()).Scan(&f.ID, &f.CreatedAt)
}

func (r *Repository) GetAll() ([]Folder, error) {
	query := `SELECT id, name, created_by, created_at FROM folders ORDER BY id DESC`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := make([]Folder, 0)
	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.ID, &f.Name, &f.CreatedBy, &f.CreatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *Repository) Delete(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM folders WHERE id = $1`, id)
	return err
}
