package folder

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

var (
	ErrInvalidName = errors.New("invalid folder name")
)

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
func (s *Service) Update(ctx context.Context, id int64, req UpdateFolderRequest) (Folder, error) {
	name := strings.TrimSpace(req.Name)

	if id <= 0 {
		return Folder{}, fmt.Errorf("%w: id must be positive", ErrInvalidName)
	}
	if name == "" {
		return Folder{}, fmt.Errorf("%w: name is required", ErrInvalidName)
	}
	if len(name) > 255 {
		return Folder{}, fmt.Errorf("%w: name too long (max 255)", ErrInvalidName)
	}

	folder, err := s.repo.Update(ctx, id, name)
	if err != nil {
		// Don't wrap ErrFolderNotFound — let the handler map it to 404.

		return Folder{}, fmt.Errorf("service update: %w", err)
	}

	return folder, nil
}
