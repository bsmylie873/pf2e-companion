package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockItemRepository struct{ mock.Mock }

func (m *MockItemRepository) Create(item *models.Item) error {
	return m.Called(item).Error(0)
}

func (m *MockItemRepository) FindByGameID(gameID uuid.UUID) ([]models.Item, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemRepository) FindByCharacterID(characterID uuid.UUID) ([]models.Item, error) {
	args := m.Called(characterID)
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemRepository) FindByID(id uuid.UUID) (models.Item, error) {
	args := m.Called(id)
	return args.Get(0).(models.Item), args.Error(1)
}

func (m *MockItemRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Item, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.Item), args.Error(1)
}

func (m *MockItemRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}
