package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPinService struct{ mock.Mock }

func (m *MockPinService) CreatePin(sessionID, callerID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error) {
	args := m.Called(sessionID, callerID, pin)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinService) CreateGamePin(gameID, callerID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error) {
	args := m.Called(gameID, callerID, pin)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinService) ListGamePins(gameID, callerID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}

func (m *MockPinService) GetPin(id, callerID uuid.UUID) (models.SessionPin, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinService) UpdatePin(id, callerID uuid.UUID, updates map[string]interface{}) (models.SessionPin, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinService) DeletePin(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}

func (m *MockPinService) CreateMapPin(mapID, callerID uuid.UUID, pin *models.SessionPin) (models.SessionPin, error) {
	args := m.Called(mapID, callerID, pin)
	return args.Get(0).(models.SessionPin), args.Error(1)
}

func (m *MockPinService) ListMapPins(mapID, callerID uuid.UUID) ([]models.SessionPin, error) {
	args := m.Called(mapID, callerID)
	return args.Get(0).([]models.SessionPin), args.Error(1)
}
