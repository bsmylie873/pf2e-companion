package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockMembershipService struct{ mock.Mock }

func (m *MockMembershipService) CreateMembership(membership *models.GameMembership, callerID uuid.UUID) (models.GameMembership, error) {
	args := m.Called(membership, callerID)
	return args.Get(0).(models.GameMembership), args.Error(1)
}

func (m *MockMembershipService) ListMemberships(gameID, callerID uuid.UUID) ([]models.GameMembership, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).([]models.GameMembership), args.Error(1)
}

func (m *MockMembershipService) GetMembership(id, callerID uuid.UUID) (models.GameMembership, error) {
	args := m.Called(id, callerID)
	return args.Get(0).(models.GameMembership), args.Error(1)
}

func (m *MockMembershipService) UpdateMembership(id, callerID uuid.UUID, updates map[string]interface{}) (models.GameMembership, error) {
	args := m.Called(id, callerID, updates)
	return args.Get(0).(models.GameMembership), args.Error(1)
}

func (m *MockMembershipService) DeleteMembership(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}
