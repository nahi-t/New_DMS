package share

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
	ErrInvalidPermission = errors.New("permission must be 'viewer' or 'editor'")
	ErrForbidden         = errors.New("forbidden")
	ErrSelfShare         = errors.New("cannot share a document with yourself")
)

/* ------------------------------------------------------------------ */
/*  Authorization                                                      */
/* ------------------------------------------------------------------ */

// CanShare reports whether the actor may manage shares for a document.
//
//   - admin   → always
//   - manager → always
//   - folder creator → yes, for documents inside their folder
//   - document uploader → yes, for their own document
//   - everyone else → no
func (s *Service) CanShare(
	ctx context.Context,
	documentID, actorID int64,
	actorRole string,
) (bool, error) {
	role := strings.ToLower(actorRole)
	if role == "admin" || role == "manager" {
		return true, nil
	}

	// Document uploader?
	docOwner, err := s.repo.GetDocumentOwner(ctx, documentID)
	if err == nil && docOwner == actorID {
		return true, nil
	}

	// Folder creator?
	folderID, err := s.repo.GetDocumentFolderID(ctx, documentID)
	if err != nil {
		return false, err
	}
	folderOwner, err := s.repo.GetFolderOwner(ctx, folderID)
	if err != nil {
		return false, err
	}
	return folderOwner == actorID, nil
}

/* ------------------------------------------------------------------ */
/*  Mutations                                                          */
/* ------------------------------------------------------------------ */

// Share grants a user access to a document.
func (s *Service) Share(
	ctx context.Context,
	documentID, userID, sharedBy int64,
	actorRole string,
	permission string,
) error {
	if permission != "viewer" && permission != "editor" {
		return ErrInvalidPermission
	}
	if documentID <= 0 || userID <= 0 {
		return fmt.Errorf("invalid document or user id")
	}
	if userID == sharedBy {
		return ErrSelfShare
	}

	ok, err := s.CanShare(ctx, documentID, sharedBy, actorRole)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	return s.repo.Share(ctx, documentID, userID, sharedBy, permission)
}

// Unshare removes a user's access to a document.
func (s *Service) Unshare(
	ctx context.Context,
	documentID, userID, actorID int64,
	actorRole string,
) error {
	ok, err := s.CanShare(ctx, documentID, actorID, actorRole)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return s.repo.Unshare(ctx, documentID, userID)
}

/* ------------------------------------------------------------------ */
/*  Queries                                                            */
/* ------------------------------------------------------------------ */

// ListShares returns everyone a document is shared with.
// Only actors who can share are allowed to see the list.
func (s *Service) ListShares(
	ctx context.Context,
	documentID, actorID int64,
	actorRole string,
) ([]DocumentShare, error) {
	ok, err := s.CanShare(ctx, documentID, actorID, actorRole)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	return s.repo.ListShares(ctx, documentID)
}

// GetSharedWithMe returns everything shared with the given user.
// No restriction — every authenticated user can see their own list.
func (s *Service) GetSharedWithMe(ctx context.Context, userID int64) ([]DocumentShare, error) {
	return s.repo.GetSharedWithMe(ctx, userID)
}

// CanAccess is consumed by the document package through its ShareChecker
// interface. It answers: "may this user perform `action` on this document?"
func (s *Service) CanAccess(
	ctx context.Context,
	documentID, userID int64,
	required string,
) (bool, error) {
	return s.repo.CanAccess(ctx, documentID, userID, required)
}
