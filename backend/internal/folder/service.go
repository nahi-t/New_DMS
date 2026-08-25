package folder

import "errors"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateFolder(name string, userID int64) (*Folder, error) {
	if name == "" {
		return nil, errors.New("folder name cannot be empty")
	}

	f := &Folder{
		Name:      name,
		CreatedBy: userID,
	}

	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) ListFolders() ([]Folder, error) {
	return s.repo.GetAll()
}

func (s *Service) DeleteFolder(id int64) error {
	return s.repo.Delete(id)
}
