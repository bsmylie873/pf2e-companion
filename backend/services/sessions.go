package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type SessionService interface {
	CreateSession(gameID, userID uuid.UUID, session *models.Session) (models.Session, error)
	ListGameSessions(gameID, userID uuid.UUID) ([]models.Session, error)
	GetSession(id, userID uuid.UUID) (models.Session, error)
	UpdateSession(id, userID uuid.UUID, updates map[string]interface{}) (models.Session, error)
	DeleteSession(id, userID uuid.UUID) error
}

type sessionService struct {
	repo           repositories.SessionRepository
	membershipRepo repositories.MembershipRepository
}

func NewSessionService(repo repositories.SessionRepository, membershipRepo repositories.MembershipRepository) SessionService {
	return &sessionService{repo: repo, membershipRepo: membershipRepo}
}

func (s *sessionService) CreateSession(gameID, userID uuid.UUID, session *models.Session) (models.Session, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.Session{}, ErrForbidden
	}
	session.ID = uuid.Nil
	session.GameID = gameID
	if err := s.repo.Create(session); err != nil {
		return models.Session{}, err
	}
	return *session, nil
}

func (s *sessionService) ListGameSessions(gameID, userID uuid.UUID) ([]models.Session, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *sessionService) GetSession(id, userID uuid.UUID) (models.Session, error) {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return models.Session{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return models.Session{}, ErrForbidden
	}
	return session, nil
}

func (s *sessionService) UpdateSession(id, userID uuid.UUID, updates map[string]interface{}) (models.Session, error) {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return models.Session{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return models.Session{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")

	var expectedVersion *int
	if raw, ok := updates["version"]; ok {
		v := int(raw.(float64))
		expectedVersion = &v
		delete(updates, "version")
	}

	return s.repo.Update(id, updates, expectedVersion)
}

func (s *sessionService) DeleteSession(id, userID uuid.UUID) error {
	session, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID); err != nil {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
