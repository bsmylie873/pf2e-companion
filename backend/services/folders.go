package services

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type FolderService interface {
	CreateFolder(gameID, userID uuid.UUID, name, folderType, visibility string) (models.Folder, error)
	ListFolders(gameID, userID uuid.UUID, folderType string) ([]models.Folder, error)
	RenameFolder(folderID, userID uuid.UUID, name string) (models.Folder, error)
	DeleteFolder(folderID, userID uuid.UUID) error
	ReorderFolders(gameID, userID uuid.UUID, folderType string, orderedIDs []uuid.UUID) error
}

type folderService struct {
	repo           repositories.FolderRepository
	membershipRepo repositories.MembershipRepository
}

func NewFolderService(repo repositories.FolderRepository, membershipRepo repositories.MembershipRepository) FolderService {
	return &folderService{repo: repo, membershipRepo: membershipRepo}
}

func (s *folderService) CreateFolder(gameID, userID uuid.UUID, name, folderType, visibility string) (models.Folder, error) {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, gameID)
	if err != nil {
		return models.Folder{}, ErrForbidden
	}

	if folderType == "session" && !membership.IsGM {
		return models.Folder{}, ErrSessionFoldersReadOnly
	}

	if folderType != "session" && folderType != "note" {
		return models.Folder{}, fmt.Errorf("%w: folder_type must be 'session' or 'note'", ErrValidation)
	}

	if visibility != "private" && visibility != "game-wide" {
		return models.Folder{}, fmt.Errorf("%w: visibility must be 'private' or 'game-wide'", ErrValidation)
	}

	if len(name) == 0 || len(name) > 100 {
		return models.Folder{}, fmt.Errorf("%w: name must be 1-100 characters", ErrValidation)
	}

	folder := &models.Folder{
		GameID:     gameID,
		Name:       name,
		FolderType: folderType,
		Visibility: visibility,
	}

	if folderType == "session" {
		folder.UserID = nil
	} else {
		folder.UserID = &userID
	}

	// Assign next position
	maxPos, err := s.repo.MaxPosition(gameID, folderType, folder.UserID)
	if err != nil {
		return models.Folder{}, err
	}
	folder.Position = maxPos + 1

	if err := s.repo.Create(folder); err != nil {
		if strings.Contains(err.Error(), "23505") {
			return models.Folder{}, ErrConflict
		}
		return models.Folder{}, err
	}

	return *folder, nil
}

func (s *folderService) ListFolders(gameID, userID uuid.UUID, folderType string) ([]models.Folder, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}

	var folders []models.Folder
	var err error

	if folderType == "session" {
		folders, err = s.repo.FindSessionFolders(gameID)
	} else {
		folders, err = s.repo.FindNoteFolders(gameID, userID)
	}
	if err != nil {
		return nil, err
	}

	// Filter out private folders not owned by the requesting user
	result := make([]models.Folder, 0, len(folders))
	for _, f := range folders {
		if f.Visibility == "private" && (f.UserID == nil || *f.UserID != userID) {
			continue
		}
		result = append(result, f)
	}

	return result, nil
}

func (s *folderService) RenameFolder(folderID, userID uuid.UUID, name string) (models.Folder, error) {
	folder, err := s.repo.FindByID(folderID)
	if err != nil {
		return models.Folder{}, err
	}

	if err := s.checkFolderPermission(folder, userID); err != nil {
		return models.Folder{}, err
	}

	if len(name) == 0 || len(name) > 100 {
		return models.Folder{}, fmt.Errorf("%w: name must be 1-100 characters", ErrValidation)
	}

	updated, err := s.repo.Update(folderID, map[string]interface{}{"name": name})
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return models.Folder{}, ErrConflict
		}
		return models.Folder{}, err
	}

	return updated, nil
}

func (s *folderService) DeleteFolder(folderID, userID uuid.UUID) error {
	folder, err := s.repo.FindByID(folderID)
	if err != nil {
		return err
	}

	if err := s.checkFolderPermission(folder, userID); err != nil {
		return err
	}

	// ON DELETE SET NULL handles unfiling of sessions and notes at the DB level
	return s.repo.Delete(folderID)
}

func (s *folderService) ReorderFolders(gameID, userID uuid.UUID, folderType string, orderedIDs []uuid.UUID) error {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, gameID)
	if err != nil {
		return ErrForbidden
	}

	if folderType == "session" && !membership.IsGM {
		return ErrSessionFoldersReadOnly
	}

	// Load all folders in scope to validate IDs
	var scopeFolders []models.Folder
	if folderType == "session" {
		scopeFolders, err = s.repo.FindSessionFolders(gameID)
	} else {
		scopeFolders, err = s.repo.FindNoteFolders(gameID, userID)
	}
	if err != nil {
		return err
	}

	scopeMap := make(map[uuid.UUID]bool, len(scopeFolders))
	for _, f := range scopeFolders {
		scopeMap[f.ID] = true
	}

	// Validate all provided IDs belong to the scope
	for _, id := range orderedIDs {
		if !scopeMap[id] {
			return ErrForbidden
		}
	}

	positions := make([]int, len(orderedIDs))
	for i := range orderedIDs {
		positions[i] = i
	}

	return s.repo.BatchUpdatePositions(orderedIDs, positions)
}

func (s *folderService) checkFolderPermission(folder models.Folder, userID uuid.UUID) error {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, folder.GameID)
	if err != nil {
		return ErrForbidden
	}

	if folder.FolderType == "session" {
		if !membership.IsGM {
			return ErrSessionFoldersReadOnly
		}
		return nil
	}

	// Note folders: must be the owner
	if folder.UserID == nil || *folder.UserID != userID {
		return ErrForbidden
	}

	return nil
}
