package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockMembershipRepository struct{ mock.Mock }

func (m *MockMembershipRepository) Create(membership *models.GameMembership) error {
	return m.Called(membership).Error(0)
}

func (m *MockMembershipRepository) FindByGameID(gameID uuid.UUID) ([]models.GameMembership, error) {
	args := m.Called(gameID)
	return args.Get(0).([]models.GameMembership), args.Error(1)
}

func (m *MockMembershipRepository) FindByID(id uuid.UUID) (models.GameMembership, error) {
	args := m.Called(id)
	return args.Get(0).(models.GameMembership), args.Error(1)
}

func (m *MockMembershipRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.GameMembership, error) {
	args := m.Called(id, updates)
	return args.Get(0).(models.GameMembership), args.Error(1)
}

func (m *MockMembershipRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockMembershipRepository) FindByUserID(userID uuid.UUID) ([]models.GameMembership, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.GameMembership), args.Error(1)
}

func (m *MockMembershipRepository) FindByUserAndGameID(userID, gameID uuid.UUID) (models.GameMembership, error) {
	args := m.Called(userID, gameID)
	return args.Get(0).(models.GameMembership), args.Error(1)
}
