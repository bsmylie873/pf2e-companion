package services

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

// MapService defines business logic for GameMap operations.
type MapService interface {
	CreateMap(gameID, userID uuid.UUID, name string, description *string) (models.GameMap, error)
	ListMaps(gameID, userID uuid.UUID) ([]models.GameMap, error)
	ListArchivedMaps(gameID, userID uuid.UUID) ([]models.GameMap, error)
	GetMap(mapID, userID uuid.UUID) (models.GameMap, error)
	RenameMap(mapID, userID uuid.UUID, name string, description *string) (models.GameMap, error)
	ReorderMaps(gameID, userID uuid.UUID, mapIDs []uuid.UUID) error
	ArchiveMap(mapID, userID uuid.UUID) error
	RestoreMap(mapID, userID uuid.UUID) (models.GameMap, error)
	SetMapImage(mapID, userID uuid.UUID, imageURL string) (models.GameMap, error)
	DeleteMapImage(mapID, userID uuid.UUID) (models.GameMap, error)
	CleanupExpiredMaps() error
}

type mapService struct {
	repo           repositories.MapRepository
	membershipRepo repositories.MembershipRepository
}

// NewMapService constructs a MapService.
func NewMapService(repo repositories.MapRepository, membershipRepo repositories.MembershipRepository) MapService {
	return &mapService{repo: repo, membershipRepo: membershipRepo}
}

func (s *mapService) requireGM(userID, gameID uuid.UUID) error {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, gameID)
	if err != nil {
		return ErrForbidden
	}
	if !membership.IsGM {
		return ErrForbidden
	}
	return nil
}

func (s *mapService) CreateMap(gameID, userID uuid.UUID, name string, description *string) (models.GameMap, error) {
	if err := s.requireGM(userID, gameID); err != nil {
		return models.GameMap{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return models.GameMap{}, ErrValidation
	}
	// Compute next sort_order
	existing, err := s.repo.FindActiveByGameID(gameID)
	if err != nil {
		return models.GameMap{}, err
	}
	sortOrder := len(existing)

	m := &models.GameMap{
		GameID:      gameID,
		Name:        name,
		Description: description,
		SortOrder:   sortOrder,
	}
	if err := s.repo.Create(m); err != nil {
		if isUniqueViolation(err) {
			return models.GameMap{}, ErrConflict
		}
		return models.GameMap{}, err
	}
	return *m, nil
}

func (s *mapService) ListMaps(gameID, userID uuid.UUID) ([]models.GameMap, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindActiveByGameID(gameID)
}

func (s *mapService) ListArchivedMaps(gameID, userID uuid.UUID) ([]models.GameMap, error) {
	if err := s.requireGM(userID, gameID); err != nil {
		return nil, err
	}
	return s.repo.FindArchivedByGameID(gameID)
}

func (s *mapService) GetMap(mapID, userID uuid.UUID) (models.GameMap, error) {
	m, err := s.repo.FindByID(mapID)
	if err != nil {
		return models.GameMap{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, m.GameID); err != nil {
		return models.GameMap{}, ErrForbidden
	}
	return m, nil
}

func (s *mapService) RenameMap(mapID, userID uuid.UUID, name string, description *string) (models.GameMap, error) {
	m, err := s.repo.FindByID(mapID)
	if err != nil {
		return models.GameMap{}, err
	}
	if err := s.requireGM(userID, m.GameID); err != nil {
		return models.GameMap{}, err
	}
	updates := map[string]interface{}{}
	name = strings.TrimSpace(name)
	if name != "" {
		updates["name"] = name
	}
	if description != nil {
		updates["description"] = *description
	}
	if len(updates) == 0 {
		return m, nil
	}
	updated, err := s.repo.Update(mapID, updates)
	if err != nil {
		if isUniqueViolation(err) {
			return models.GameMap{}, ErrConflict
		}
		return models.GameMap{}, err
	}
	return updated, nil
}

func (s *mapService) ReorderMaps(gameID, userID uuid.UUID, mapIDs []uuid.UUID) error {
	if err := s.requireGM(userID, gameID); err != nil {
		return err
	}
	for i, id := range mapIDs {
		if _, err := s.repo.Update(id, map[string]interface{}{"sort_order": i}); err != nil {
			return err
		}
	}
	return nil
}

func (s *mapService) ArchiveMap(mapID, userID uuid.UUID) error {
	m, err := s.repo.FindByID(mapID)
	if err != nil {
		return err
	}
	if err := s.requireGM(userID, m.GameID); err != nil {
		return err
	}
	return s.repo.Archive(mapID)
}

func (s *mapService) RestoreMap(mapID, userID uuid.UUID) (models.GameMap, error) {
	m, err := s.repo.FindByID(mapID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.GameMap{}, gorm.ErrRecordNotFound
		}
		return models.GameMap{}, err
	}
	if err := s.requireGM(userID, m.GameID); err != nil {
		return models.GameMap{}, err
	}
	if m.ArchivedAt == nil {
		return models.GameMap{}, gorm.ErrRecordNotFound
	}
	// Check 24-hour window
	if time.Since(*m.ArchivedAt) > 24*time.Hour {
		return models.GameMap{}, gorm.ErrRecordNotFound
	}
	updated, err := s.repo.Update(mapID, map[string]interface{}{"archived_at": nil})
	if err != nil {
		return models.GameMap{}, err
	}
	return updated, nil
}

func (s *mapService) SetMapImage(mapID, userID uuid.UUID, imageURL string) (models.GameMap, error) {
	m, err := s.repo.FindByID(mapID)
	if err != nil {
		return models.GameMap{}, err
	}
	if err := s.requireGM(userID, m.GameID); err != nil {
		return models.GameMap{}, err
	}
	// Delete old image file if present
	if m.ImageURL != nil {
		_ = os.Remove("." + *m.ImageURL)
	}
	return s.repo.Update(mapID, map[string]interface{}{"image_url": imageURL})
}

func (s *mapService) DeleteMapImage(mapID, userID uuid.UUID) (models.GameMap, error) {
	m, err := s.repo.FindByID(mapID)
	if err != nil {
		return models.GameMap{}, err
	}
	if err := s.requireGM(userID, m.GameID); err != nil {
		return models.GameMap{}, err
	}
	if m.ImageURL != nil {
		_ = os.Remove("." + *m.ImageURL)
	}
	return s.repo.Update(mapID, map[string]interface{}{"image_url": nil})
}

func (s *mapService) CleanupExpiredMaps() error {
	cutoff := time.Now().Add(-24 * time.Hour)
	maps, err := s.repo.FindExpiredArchived(cutoff)
	if err != nil {
		return err
	}
	for _, m := range maps {
		if m.ImageURL != nil {
			_ = os.Remove("." + *m.ImageURL)
		}
		if err := s.repo.HardDelete(m.ID); err != nil {
			return fmt.Errorf("hard delete map %s: %w", m.ID, err)
		}
	}
	return nil
}

// isUniqueViolation checks if an error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "duplicate key")
}
