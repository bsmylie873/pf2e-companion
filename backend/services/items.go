package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type ItemService interface {
	CreateItem(gameID uuid.UUID, item *models.Item) (models.Item, error)
	ListGameItems(gameID uuid.UUID) ([]models.Item, error)
	ListCharacterItems(characterID uuid.UUID) ([]models.Item, error)
	GetItem(id uuid.UUID) (models.Item, error)
	UpdateItem(id uuid.UUID, updates map[string]interface{}) (models.Item, error)
	DeleteItem(id uuid.UUID) error
}

type itemService struct {
	repo repositories.ItemRepository
}

func NewItemService(repo repositories.ItemRepository) ItemService {
	return &itemService{repo: repo}
}

func (s *itemService) CreateItem(gameID uuid.UUID, item *models.Item) (models.Item, error) {
	item.ID = uuid.Nil
	item.GameID = gameID
	if err := s.repo.Create(item); err != nil {
		return models.Item{}, err
	}
	return *item, nil
}

func (s *itemService) ListGameItems(gameID uuid.UUID) ([]models.Item, error) {
	return s.repo.FindByGameID(gameID)
}

func (s *itemService) ListCharacterItems(characterID uuid.UUID) ([]models.Item, error) {
	return s.repo.FindByCharacterID(characterID)
}

func (s *itemService) GetItem(id uuid.UUID) (models.Item, error) {
	return s.repo.FindByID(id)
}

func (s *itemService) UpdateItem(id uuid.UUID, updates map[string]interface{}) (models.Item, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.Item{}, err
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *itemService) DeleteItem(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
