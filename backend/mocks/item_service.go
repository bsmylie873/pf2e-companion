package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockItemService struct{ mock.Mock }

func (m *MockItemService) CreateItem(gameID, callerID uuid.UUID, item *models.Item) (models.Item, error) {
	args := m.Called(gameID, callerID, item)
	return args.Get(0).(models.Item), args.Error(1)
}

func (m *MockItemService) ListGameItems(gameID, callerID uuid.UUID) ([]models.Item, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemService) ListCharacterItems(characterID, callerID uuid.UUID) ([]models.Item, error) {
	args := m.Called(characterID, callerID)
	return args.Get(0).([]models.Item), args.Error(1)
}

func (m *MockItemService) GetItem(id, callerID uuid.UUID) (models.Item, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.Item), args.Error(1)
}

func (m *MockItemService) UpdateItem(id, callerID uuid.UUID, updates map[string]interface{}) (models.Item, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.Item), args.Error(1)
}

func (m *MockItemService) DeleteItem(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}
