package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type CharacterService interface {
	CreateCharacter(gameID, userID uuid.UUID, character *models.Character) (models.Character, error)
	ListGameCharacters(gameID, userID uuid.UUID) ([]models.Character, error)
	GetCharacter(id, userID uuid.UUID) (models.Character, error)
	UpdateCharacter(id, userID uuid.UUID, updates map[string]interface{}) (models.Character, error)
	DeleteCharacter(id, userID uuid.UUID) error
}

type characterService struct {
	repo           repositories.CharacterRepository
	membershipRepo repositories.MembershipRepository
}

func NewCharacterService(repo repositories.CharacterRepository, membershipRepo repositories.MembershipRepository) CharacterService {
	return &characterService{repo: repo, membershipRepo: membershipRepo}
}

func (s *characterService) CreateCharacter(gameID, userID uuid.UUID, character *models.Character) (models.Character, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.Character{}, ErrForbidden
	}
	character.ID = uuid.Nil
	character.GameID = gameID
	if err := s.repo.Create(character); err != nil {
		return models.Character{}, err
	}
	return *character, nil
}

func (s *characterService) ListGameCharacters(gameID, userID uuid.UUID) ([]models.Character, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *characterService) GetCharacter(id, userID uuid.UUID) (models.Character, error) {
	char, err := s.repo.FindByID(id)
	if err != nil {
		return models.Character{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, char.GameID); err != nil {
		return models.Character{}, ErrForbidden
	}
	return char, nil
}

func (s *characterService) UpdateCharacter(id, userID uuid.UUID, updates map[string]interface{}) (models.Character, error) {
	char, err := s.repo.FindByID(id)
	if err != nil {
		return models.Character{}, err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, char.GameID); err != nil {
		return models.Character{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *characterService) DeleteCharacter(id, userID uuid.UUID) error {
	char, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, char.GameID); err != nil {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
