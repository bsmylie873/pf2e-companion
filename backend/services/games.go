package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type GameService interface {
	CreateGame(game *models.Game, members []models.GameMembership, creatorID uuid.UUID) (models.Game, error)
	ListGames(userID uuid.UUID) ([]models.GameWithRole, error)
	ListGamesPaginated(userID uuid.UUID, offset, limit int) ([]models.GameWithRole, int64, error)
	GetGame(id, userID uuid.UUID) (models.Game, error)
	UpdateGame(id, userID uuid.UUID, updates map[string]interface{}) (models.Game, error)
	DeleteGame(id, userID uuid.UUID) error
}

type gameService struct {
	repo           repositories.GameRepository
	membershipRepo repositories.MembershipRepository
}

func NewGameService(repo repositories.GameRepository, membershipRepo repositories.MembershipRepository) GameService {
	return &gameService{repo: repo, membershipRepo: membershipRepo}
}

func (s *gameService) CreateGame(game *models.Game, members []models.GameMembership, creatorID uuid.UUID) (models.Game, error) {
	if err := s.repo.Create(game); err != nil {
		return models.Game{}, err
	}
	if len(members) > 0 {
		for i := range members {
			members[i].GameID = game.ID
			if err := s.membershipRepo.Create(&members[i]); err != nil {
				return models.Game{}, err
			}
		}
	} else {
		membership := &models.GameMembership{GameID: game.ID, UserID: creatorID, IsGM: true}
		if err := s.membershipRepo.Create(membership); err != nil {
			return models.Game{}, err
		}
	}
	return *game, nil
}

func (s *gameService) ListGames(userID uuid.UUID) ([]models.GameWithRole, error) {
	memberships, err := s.membershipRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(memberships))
	gmMap := make(map[uuid.UUID]bool)
	for i, m := range memberships {
		ids[i] = m.GameID
		gmMap[m.GameID] = m.IsGM
	}
	games, err := s.repo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	result := make([]models.GameWithRole, len(games))
	for i, g := range games {
		result[i] = models.GameWithRole{Game: g, IsGM: gmMap[g.ID]}
	}
	return result, nil
}

func (s *gameService) ListGamesPaginated(userID uuid.UUID, offset, limit int) ([]models.GameWithRole, int64, error) {
	memberships, err := s.membershipRepo.FindByUserID(userID)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]uuid.UUID, len(memberships))
	gmMap := make(map[uuid.UUID]bool)
	for i, m := range memberships {
		ids[i] = m.GameID
		gmMap[m.GameID] = m.IsGM
	}
	games, total, err := s.repo.FindByIDsPaginated(ids, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	result := make([]models.GameWithRole, len(games))
	for i, g := range games {
		result[i] = models.GameWithRole{Game: g, IsGM: gmMap[g.ID]}
	}
	return result, total, nil
}

func (s *gameService) GetGame(id, userID uuid.UUID) (models.Game, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, id); err != nil {
		return models.Game{}, ErrForbidden
	}
	return s.repo.FindByID(id)
}

func (s *gameService) UpdateGame(id, userID uuid.UUID, updates map[string]interface{}) (models.Game, error) {
	m, err := s.membershipRepo.FindByUserAndGameID(userID, id)
	if err != nil {
		return models.Game{}, ErrForbidden
	}
	if !m.IsGM {
		return models.Game{}, ErrForbidden
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *gameService) DeleteGame(id, userID uuid.UUID) error {
	m, err := s.membershipRepo.FindByUserAndGameID(userID, id)
	if err != nil {
		return ErrForbidden
	}
	if !m.IsGM {
		return ErrForbidden
	}
	return s.repo.Delete(id)
}
