package folder

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateFolder(ctx context.Context, name string, userID int64) (*Folder, error) {
	if name == "" {
		return nil, errors.New("folder name cannot be empty")
	}

	f := &Folder{
		Name:      name,
		CreatedBy: userID,
	}

	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) ListFolders(ctx context.Context) ([]Folder, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) DeleteFolder(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
