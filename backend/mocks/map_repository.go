package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockMapRepository struct{ mock.Mock }

func (m *MockMapRepository) Create(gameMap *models.GameMap) error {
	return m.Called(gameMap).Error(0)
}

func (m *MockMapRepository) FindByID(id uuid.UUID) (models.GameMap, error) {
	args := m.Called(id)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapRepository) FindByGameID(gameID uuid.UUID) ([]models.GameMap, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.GameMap), args.Error(1)
}

func (m *MockMapRepository) FindActiveByGameID(gameID uuid.UUID) ([]models.GameMap, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.GameMap), args.Error(1)
}

func (m *MockMapRepository) FindArchivedByGameID(gameID uuid.UUID) ([]models.GameMap, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.GameMap), args.Error(1)
}

func (m *MockMapRepository) FindByGameIDAndName(gameID uuid.UUID, name string) (models.GameMap, error) {
	args := m.Called(gameID, name)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.GameMap, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.GameMap), args.Error(1)
}

func (m *MockMapRepository) Archive(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockMapRepository) HardDelete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockMapRepository) FindExpiredArchived(before time.Time) ([]models.GameMap, error) {
	args := m.Called(before)
	return args.Get(0).([]models.GameMap), args.Error(1)
}
