package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockGameService struct{ mock.Mock }

func (m *MockGameService) CreateGame(game *models.Game, members []models.GameMembership, creatorID uuid.UUID) (models.Game, error) {
	args := m.Called(game, members, creatorID)
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameService) ListGames(userID uuid.UUID) ([]models.GameWithRole, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.GameWithRole), args.Error(1)
}

func (m *MockGameService) ListGamesPaginated(userID uuid.UUID, offset, limit int) ([]models.GameWithRole, int64, error) {
	args := m.Called(userID, offset, limit)
	return args.Get(0).([]models.GameWithRole), args.Get(1).(int64), args.Error(2)
}

func (m *MockGameService) GetGame(id, userID uuid.UUID) (models.Game, error) {
	args := m.Called(id, userID)
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameService) UpdateGame(id, userID uuid.UUID, updates map[string]interface{}) (models.Game, error) {
	args := m.Called(id, userID, updates)
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameService) DeleteGame(id, userID uuid.UUID) error {
	return m.Called(id, userID).Error(0)
}
