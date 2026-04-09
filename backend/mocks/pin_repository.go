package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPinRepository struct{ mock.Mock }

func (m *MockPinRepository) Create(pin *models.SessionPin) error {
	return m.Called(pin).Error(0)
}

func (m *MockPinRepository) FindByGameID(gameID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}

func (m *MockPinRepository) FindByID(id uuid.UUID) (models.SessionPin, error) {
	args := m.Called(id)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.SessionPin, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockPinRepository) FindByGroupID(groupID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(groupID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}

func (m *MockPinRepository) ClearGroupID(groupID uuid.UUID) error {
	return m.Called(groupID).Error(0)
}

func (m *MockPinRepository) SetGroupID(pinID, groupID uuid.UUID) error {
	return m.Called(pinID, groupID).Error(0)
}

func (m *MockPinRepository) FindStandaloneByGameID(gameID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}

func (m *MockPinRepository) FindByMapID(mapID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(mapID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}

func (m *MockPinRepository) FindStandaloneByMapID(mapID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(mapID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}
