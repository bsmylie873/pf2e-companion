package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockUserService struct{ mock.Mock }

func (m *MockUserService) ListUsers() ([]models.UserPublicResponse, error) {
	args := m.Called()
	return args.Get(0).([]models.UserPublicResponse), args.Error(1)
}

func (m *MockUserService) GetUser(id uuid.UUID) (models.UserResponse, error) {
	args := m.Called(id)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(id uuid.UUID, updates map[string]interface{}, callerID uuid.UUID) (models.UserResponse, error) {
	args := m.Called(id, updates, callerID)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(id, callerID uuid.UUID) error {
	return m.Called(id, callerID).Error(0)
}
