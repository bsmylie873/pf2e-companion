package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockCharacterService struct{ mock.Mock }

func (m *MockCharacterService) CreateCharacter(gameID, callerID uuid.UUID, character *models.Character) (models.Character, error) {
	args := m.Called(gameID, callerID, character)
	return args.Get(0).(models.Character), args.Error(1)
}

func (m *MockCharacterService) ListGameCharacters(gameID, callerID uuid.UUID) ([]models.Character, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.Character), args.Error(1)
}

func (m *MockCharacterService) GetCharacter(id, callerID uuid.UUID) (models.Character, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.Character), args.Error(1)
}

func (m *MockCharacterService) UpdateCharacter(id, callerID uuid.UUID, updates map[string]interface{}) (models.Character, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.Character), args.Error(1)
}

func (m *MockCharacterService) DeleteCharacter(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}
