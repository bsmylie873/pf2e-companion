package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type PinService interface {
	CreatePin(sessionID, userID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error)
	CreateGamePin(gameID, userID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error)
	ListGamePins(gameID, userID uuid.UUID) ([]models.SessionPin, error)
	GetPin(id, userID uuid.UUID) (models.SessionPin, error)
	UpdatePin(id, userID uuid.UUID, updates map[string]interface{}) (models.SessionPin, error)
	DeletePin(id, userID uuid.UUID) error
}

type pinService struct {
	repo           repositories.PinRepository
	sessionRepo    repositories.SessionRepository
	membershipRepo repositories.MembershipRepository
	pinGroupRepo   repositories.PinGroupRepository
}

func NewPinService(repo repositories.PinRepository, sessionRepo repositories.SessionRepository, membershipRepo repositories.MembershipRepository, pinGroupRepo repositories.PinGroupRepository) PinService {
	return &pinService{repo: repo, sessionRepo: sessionRepo, membershipRepo: membershipRepo, pinGroupRepo: pinGroupRepo}
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
	pin.GameID = session.GameID
	pin.SessionID = &sessionID
	if err := s.repo.Create(pin); err != nil {
		return models.SessionPin{}, err
	}
	return *pin, nil
}

func (s *pinService) CreateGamePin(gameID, userID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.SessionPin{}, ErrForbidden
	}
	pin.ID = uuid.Nil
	pin.GameID = gameID
	pin.SessionID = nil
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
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, pin.GameID); err != nil {
		return models.SessionPin{}, ErrForbidden
	}
	return pin, nil
}

func (s *pinService) UpdatePin(id, userID uuid.UUID, updates map[string]interface{}) (models.SessionPin, error) {
	pin, err := s.repo.FindByID(id)
	if err != nil {
		return models.SessionPin{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, pin.GameID); err != nil {
		return models.SessionPin{}, ErrForbidden
	}
	// Prevent repositioning pins that belong to a group
	if pin.GroupID != nil {
		if _, hasX := updates["x"]; hasX {
			return models.SessionPin{}, ErrGroupedPinMove
		}
		if _, hasY := updates["y"]; hasY {
			return models.SessionPin{}, ErrGroupedPinMove
		}
	}
	delete(updates, "id")
	delete(updates, "game_id")
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
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, pin.GameID); err != nil {
		return ErrForbidden
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	// Auto-dissolve group if this pin was the last/only other member
	if pin.GroupID != nil {
		count, err := s.pinGroupRepo.CountMembers(*pin.GroupID)
		if err != nil {
			return err
		}
		if count <= 1 {
			if count == 1 {
				remaining, err := s.repo.FindByGroupID(*pin.GroupID)
				if err != nil {
					return err
				}
				for _, p := range remaining {
					if err := s.repo.ClearGroupID(p.ID); err != nil {
						return err
					}
				}
			}
			if err := s.pinGroupRepo.Delete(*pin.GroupID); err != nil {
				return err
			}
		}
	}
	return nil
}
