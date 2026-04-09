package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockGameRepository struct{ mock.Mock }

func (m *MockGameRepository) Create(game *models.Game) error {
	return m.Called(game).Error(0)
}

func (m *MockGameRepository) FindAll() ([]models.Game, error) {
	args := m.Called()
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockGameRepository) FindByID(id uuid.UUID) (models.Game, error) {
	args := m.Called(id)
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Game, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockGameRepository) FindByIDs(ids []uuid.UUID) ([]models.Game, error) {
	args := m.Called(ids)
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockGameRepository) FindByIDsPaginated(ids []uuid.UUID, offset, limit int) ([]models.Game, int64, error) {
	args := m.Called(ids, offset, limit)
	return args.Get(0).([]models.Game), args.Get(1).(int64), args.Error(2)
}
