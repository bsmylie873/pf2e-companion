package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type MembershipService interface {
	CreateMembership(membership *models.GameMembership) (models.GameMembership, error)
	ListMemberships(gameID uuid.UUID) ([]models.GameMembership, error)
	GetMembership(id uuid.UUID) (models.GameMembership, error)
	UpdateMembership(id uuid.UUID, updates map[string]interface{}) (models.GameMembership, error)
	DeleteMembership(id uuid.UUID) error
}

type membershipService struct {
	repo repositories.MembershipRepository
}

func NewMembershipService(repo repositories.MembershipRepository) MembershipService {
	return &membershipService{repo: repo}
}

func (s *membershipService) CreateMembership(membership *models.GameMembership) (models.GameMembership, error) {
	if err := s.repo.Create(membership); err != nil {
		return models.GameMembership{}, err
	}
	return *membership, nil
}

func (s *membershipService) ListMemberships(gameID uuid.UUID) ([]models.GameMembership, error) {
	return s.repo.FindByGameID(gameID)
}

func (s *membershipService) GetMembership(id uuid.UUID) (models.GameMembership, error) {
	return s.repo.FindByID(id)
}

func (s *membershipService) UpdateMembership(id uuid.UUID, updates map[string]interface{}) (models.GameMembership, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.GameMembership{}, err
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *membershipService) DeleteMembership(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
