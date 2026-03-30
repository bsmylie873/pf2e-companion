package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type MembershipService interface {
	CreateMembership(membership *models.GameMembership, callerID uuid.UUID) (models.GameMembership, error)
	ListMemberships(gameID, callerID uuid.UUID) ([]models.GameMembership, error)
	GetMembership(id, callerID uuid.UUID) (models.GameMembership, error)
	UpdateMembership(id, callerID uuid.UUID, updates map[string]interface{}) (models.GameMembership, error)
	DeleteMembership(id, callerID uuid.UUID) error
}

type membershipService struct {
	repo    repositories.MembershipRepository
	prefSvc PreferenceService
}

func NewMembershipService(repo repositories.MembershipRepository, prefSvc PreferenceService) MembershipService {
	return &membershipService{repo: repo, prefSvc: prefSvc}
}

func (s *membershipService) CreateMembership(membership *models.GameMembership, callerID uuid.UUID) (models.GameMembership, error) {
	if _, err := s.repo.FindByUserAndGameID(callerID, membership.GameID); err != nil {
		return models.GameMembership{}, ErrForbidden
	}
	if err := s.repo.Create(membership); err != nil {
		return models.GameMembership{}, err
	}
	return *membership, nil
}

func (s *membershipService) ListMemberships(gameID, callerID uuid.UUID) ([]models.GameMembership, error) {
	if _, err := s.repo.FindByUserAndGameID(callerID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *membershipService) GetMembership(id, callerID uuid.UUID) (models.GameMembership, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return models.GameMembership{}, err
	}
	if _, err := s.repo.FindByUserAndGameID(callerID, m.GameID); err != nil {
		return models.GameMembership{}, ErrForbidden
	}
	return m, nil
}

func (s *membershipService) UpdateMembership(id, callerID uuid.UUID, updates map[string]interface{}) (models.GameMembership, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return models.GameMembership{}, err
	}
	if _, err := s.repo.FindByUserAndGameID(callerID, m.GameID); err != nil {
		return models.GameMembership{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *membershipService) DeleteMembership(id, callerID uuid.UUID) error {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if _, err := s.repo.FindByUserAndGameID(callerID, m.GameID); err != nil {
		return ErrForbidden
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return s.prefSvc.ClearDefaultGameForMembership(m.UserID, m.GameID)
}
