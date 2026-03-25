package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type PinService interface {
	CreatePin(sessionID, userID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error)
	ListGamePins(gameID, userID uuid.UUID) ([]models.SessionPin, error)
	GetPin(id, userID uuid.UUID) (models.SessionPin, error)
	UpdatePin(id, userID uuid.UUID, updates map[string]interface{}) (models.SessionPin, error)
	DeletePin(id, userID uuid.UUID) error
}

type pinService struct {
	repo           repositories.PinRepository
	sessionRepo    repositories.SessionRepository
	membershipRepo repositories.MembershipRepository
}

func NewPinService(repo repositories.PinRepository, sessionRepo repositories.SessionRepository, membershipRepo repositories.MembershipRepository) PinService {
	return &pinService{repo: repo, sessionRepo: sessionRepo, membershipRepo: membershipRepo}
}

func (s *pinService) CreatePin(sessionID, userID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return models.SessionPin{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return models.SessionPin{}, ErrForbidden
	}
	pin.ID = uuid.Nil
	pin.SessionID = sessionID
	if err := s.repo.Create(pin); err != nil {
		return models.SessionPin{}, err
	}
	return *pin, nil
}

func (s *pinService) ListGamePins(gameID, userID uuid.UUID) ([]models.SessionPin, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *pinService) GetPin(id, userID uuid.UUID) (models.SessionPin, error) {
	pin, err := s.repo.FindByID(id)
	if err != nil {
		return models.SessionPin{}, err
	}
	session, err := s.sessionRepo.FindByID(pin.SessionID)
	if err != nil {
		return models.SessionPin{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return models.SessionPin{}, ErrForbidden
	}
	return pin, nil
}

func (s *pinService) UpdatePin(id, userID uuid.UUID, updates map[string]interface{}) (models.SessionPin, error) {
	pin, err := s.repo.FindByID(id)
	if err != nil {
		return models.SessionPin{}, err
	}
	session, err := s.sessionRepo.FindByID(pin.SessionID)
	if err != nil {
		return models.SessionPin{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return models.SessionPin{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "session_id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *pinService) DeletePin(id, userID uuid.UUID) error {
	pin, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	session, err := s.sessionRepo.FindByID(pin.SessionID)
	if err != nil {
		return err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
