package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type SessionService interface {
	CreateSession(gameID uuid.UUID, session *models.Session) (models.Session, error)
	ListGameSessions(gameID uuid.UUID) ([]models.Session, error)
	GetSession(id uuid.UUID) (models.Session, error)
	UpdateSession(id uuid.UUID, updates map[string]interface{}) (models.Session, error)
	DeleteSession(id uuid.UUID) error
}

type sessionService struct {
	repo repositories.SessionRepository
}

func NewSessionService(repo repositories.SessionRepository) SessionService {
	return &sessionService{repo: repo}
}

func (s *sessionService) CreateSession(gameID uuid.UUID, session *models.Session) (models.Session, error) {
	session.ID = uuid.Nil
	session.GameID = gameID
	if err := s.repo.Create(session); err != nil {
		return models.Session{}, err
	}
	return *session, nil
}

func (s *sessionService) ListGameSessions(gameID uuid.UUID) ([]models.Session, error) {
	return s.repo.FindByGameID(gameID)
}

func (s *sessionService) GetSession(id uuid.UUID) (models.Session, error) {
	return s.repo.FindByID(id)
}

func (s *sessionService) UpdateSession(id uuid.UUID, updates map[string]interface{}) (models.Session, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.Session{}, err
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *sessionService) DeleteSession(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
