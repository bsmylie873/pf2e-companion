package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type CharacterService interface {
	CreateCharacter(gameID uuid.UUID, character *models.Character) (models.Character, error)
	ListGameCharacters(gameID uuid.UUID) ([]models.Character, error)
	GetCharacter(id uuid.UUID) (models.Character, error)
	UpdateCharacter(id uuid.UUID, updates map[string]interface{}) (models.Character, error)
	DeleteCharacter(id uuid.UUID) error
}

type characterService struct {
	repo repositories.CharacterRepository
}

func NewCharacterService(repo repositories.CharacterRepository) CharacterService {
	return &characterService{repo: repo}
}

func (s *characterService) CreateCharacter(gameID uuid.UUID, character *models.Character) (models.Character, error) {
	character.ID = uuid.Nil
	character.GameID = gameID
	if err := s.repo.Create(character); err != nil {
		return models.Character{}, err
	}
	return *character, nil
}

func (s *characterService) ListGameCharacters(gameID uuid.UUID) ([]models.Character, error) {
	return s.repo.FindByGameID(gameID)
}

func (s *characterService) GetCharacter(id uuid.UUID) (models.Character, error) {
	return s.repo.FindByID(id)
}

func (s *characterService) UpdateCharacter(id uuid.UUID, updates map[string]interface{}) (models.Character, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.Character{}, err
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *characterService) DeleteCharacter(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
