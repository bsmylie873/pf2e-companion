package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPreferenceService struct{ mock.Mock }

func (m *MockPreferenceService) GetPreferences(userID uuid.UUID) (models.UserPreferenceResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(models.UserPreferenceResponse), args.Error(1)
}

func (m *MockPreferenceService) UpdatePreferences(userID uuid.UUID, updates map[string]interface{}) (models.UserPreferenceResponse, error) {
	args := m.Called(userID, updates)
	return args.Get(0).(models.UserPreferenceResponse), args.Error(1)
}

func (m *MockPreferenceService) ClearDefaultGameForMembership(userID, gameID uuid.UUID) error {
	return m.Called(userID, gameID).Error(0)
}
