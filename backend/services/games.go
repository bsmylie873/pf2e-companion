package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type GameService interface {
	CreateGame(game *models.Game) (models.Game, error)
	ListGames() ([]models.Game, error)
	GetGame(id uuid.UUID) (models.Game, error)
	UpdateGame(id uuid.UUID, updates map[string]interface{}) (models.Game, error)
	DeleteGame(id uuid.UUID) error
}

type gameService struct {
	repo repositories.GameRepository
}

func NewGameService(repo repositories.GameRepository) GameService {
	return &gameService{repo: repo}
}

func (s *gameService) CreateGame(game *models.Game) (models.Game, error) {
	if err := s.repo.Create(game); err != nil {
		return models.Game{}, err
	}
	return *game, nil
}

func (s *gameService) ListGames() ([]models.Game, error) {
	return s.repo.FindAll()
}

func (s *gameService) GetGame(id uuid.UUID) (models.Game, error) {
	return s.repo.FindByID(id)
}

func (s *gameService) UpdateGame(id uuid.UUID, updates map[string]interface{}) (models.Game, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.Game{}, err
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *gameService) DeleteGame(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
