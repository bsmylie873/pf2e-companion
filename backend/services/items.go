package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type ItemService interface {
	CreateItem(gameID, userID uuid.UUID, item *models.Item) (models.Item, error)
	ListGameItems(gameID, userID uuid.UUID) ([]models.Item, error)
	ListCharacterItems(characterID, userID uuid.UUID) ([]models.Item, error)
	GetItem(id, userID uuid.UUID) (models.Item, error)
	UpdateItem(id, userID uuid.UUID, updates map[string]interface{}) (models.Item, error)
	DeleteItem(id, userID uuid.UUID) error
}

type itemService struct {
	repo           repositories.ItemRepository
	membershipRepo repositories.MembershipRepository
	characterRepo  repositories.CharacterRepository
}

func NewItemService(repo repositories.ItemRepository, membershipRepo repositories.MembershipRepository, characterRepo repositories.CharacterRepository) ItemService {
	return &itemService{repo: repo, membershipRepo: membershipRepo, characterRepo: characterRepo}
}

func (s *itemService) CreateItem(gameID, userID uuid.UUID, item *models.Item) (models.Item, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.Item{}, ErrForbidden
	}
	item.ID = uuid.Nil
	item.GameID = gameID
	if err := s.repo.Create(item); err != nil {
		return models.Item{}, err
	}
	return *item, nil
}

func (s *itemService) ListGameItems(gameID, userID uuid.UUID) ([]models.Item, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *itemService) ListCharacterItems(characterID, userID uuid.UUID) ([]models.Item, error) {
	char, err := s.characterRepo.FindByID(characterID)
	if err != nil {
		return nil, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, char.GameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByCharacterID(characterID)
}

func (s *itemService) GetItem(id, userID uuid.UUID) (models.Item, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return models.Item{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, item.GameID); err != nil {
		return models.Item{}, ErrForbidden
	}
	return item, nil
}

func (s *itemService) UpdateItem(id, userID uuid.UUID, updates map[string]interface{}) (models.Item, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return models.Item{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, item.GameID); err != nil {
		return models.Item{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *itemService) DeleteItem(id, userID uuid.UUID) error {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, item.GameID); err != nil {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
