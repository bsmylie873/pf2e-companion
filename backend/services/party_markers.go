package services

import (
	"fmt"

	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type PartyMarkerService interface {
	GetPartyMarker(gameID, userID uuid.UUID) (*models.PartyMarker, error)
	UpsertPartyMarker(gameID, userID uuid.UUID, mapID uuid.UUID, x, y float64) (*models.PartyMarker, error)
	DeletePartyMarker(gameID, userID uuid.UUID) error
}

type partyMarkerService struct {
	repo           repositories.PartyMarkerRepository
	membershipRepo repositories.MembershipRepository
}

func NewPartyMarkerService(repo repositories.PartyMarkerRepository, membershipRepo repositories.MembershipRepository) PartyMarkerService {
	return &partyMarkerService{repo: repo, membershipRepo: membershipRepo}
}

func (s *partyMarkerService) GetPartyMarker(gameID, userID uuid.UUID) (*models.PartyMarker, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *partyMarkerService) UpsertPartyMarker(gameID, userID uuid.UUID, mapID uuid.UUID, x, y float64) (*models.PartyMarker, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	if x < 0 || x > 100 {
		return nil, fmt.Errorf("x must be between 0 and 100: %w", ErrValidation)
	}
	if y < 0 || y > 100 {
		return nil, fmt.Errorf("y must be between 0 and 100: %w", ErrValidation)
	}
	marker := &models.PartyMarker{
		GameID: gameID,
		MapID:  mapID,
		X:      x,
		Y:      y,
	}
	if err := s.repo.Upsert(marker); err != nil {
		return nil, err
	}
	return marker, nil
}

func (s *partyMarkerService) DeletePartyMarker(gameID, userID uuid.UUID) error {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return ErrForbidden
	}
	return s.repo.Delete(gameID)
}
