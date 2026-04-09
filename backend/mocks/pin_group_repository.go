package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPinGroupRepository struct{ mock.Mock }

func (m *MockPinGroupRepository) Create(group *models.PinGroup) error {
	return m.Called(group).Error(0)
}

func (m *MockPinGroupRepository) FindByID(id uuid.UUID) (models.PinGroup, error) {
	args := m.Called(id)
	return args.Get(0).(models.PinGroup), args.Error(1)
}

func (m *MockPinGroupRepository) FindByGameID(gameID uuid.UUID) ([]models.PinGroup, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.PinGroup), args.Error(1)
}

func (m *MockPinGroupRepository) FindByMapID(mapID uuid.UUID) ([]models.PinGroup, error) {
	args := m.Called(mapID)
	return args.Get(0).([]models.PinGroup), args.Error(1)
}

func (m *MockPinGroupRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.PinGroup, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.PinGroup), args.Error(1)
}

func (m *MockPinGroupRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockPinGroupRepository) CountMembers(groupID uuid.UUID) (int64, error) {
	args := m.Called(groupID)
	return args.Get(0).(int64), args.Error(1)
}
