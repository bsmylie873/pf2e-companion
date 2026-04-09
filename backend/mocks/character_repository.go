package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockCharacterRepository struct{ mock.Mock }

func (m *MockCharacterRepository) Create(character *models.Character) error {
	return m.Called(character).Error(0)
}

func (m *MockCharacterRepository) FindByGameID(gameID uuid.UUID) ([]models.Character, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.Character), args.Error(1)
}

func (m *MockCharacterRepository) FindByID(id uuid.UUID) (models.Character, error) {
	args := m.Called(id)
	return args.Get(0).(models.Character), args.Error(1)
}

func (m *MockCharacterRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Character, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.Character), args.Error(1)
}

func (m *MockCharacterRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}
