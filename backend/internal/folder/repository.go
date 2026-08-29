package folder

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(ctx context.Context, f *Folder) error {
	query := `INSERT INTO folders (name, created_by, created_at) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.DB.QueryRow(ctx, query, f.Name, f.CreatedBy, time.Now().UTC()).Scan(&f.ID, &f.CreatedAt)
}

func (r *Repository) GetAll(ctx context.Context) ([]Folder, error) {
	query := `SELECT id, name, created_by, created_at FROM folders ORDER BY id DESC`
	rows, err := r.DB.Query(ctx, query)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM folders WHERE id = $1`, id)
	return err
}
