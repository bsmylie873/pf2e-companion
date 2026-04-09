package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPreferenceRepository struct{ mock.Mock }

func (m *MockPreferenceRepository) FindByUserID(userID uuid.UUID) (models.UserPreference, error) {
	args := m.Called(userID)
	return args.Get(0).(models.UserPreference), args.Error(1)
}

func (m *MockPreferenceRepository) Upsert(pref *models.UserPreference) error {
	return m.Called(pref).Error(0)
}

func (m *MockPreferenceRepository) ClearDefaultGameForGame(gameID uuid.UUID) error {
	return m.Called(gameID).Error(0)
}

func (m *MockPreferenceRepository) ClearDefaultGameForMembership(userID, gameID uuid.UUID) error {
	return m.Called(userID, gameID).Error(0)
}
