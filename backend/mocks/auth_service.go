package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockAuthService struct{ mock.Mock }

func (m *MockAuthService) Register(req models.RegisterRequest) (models.UserResponse, models.TokenPair, error) {
	args := m.Called(req)
	return args.Get(0).(models.UserResponse), args.Get(1).(models.TokenPair), args.Error(2)
}

func (m *MockAuthService) Login(req models.LoginRequest) (models.UserResponse, models.TokenPair, error) {
	args := m.Called(req)
	return args.Get(0).(models.UserResponse), args.Get(1).(models.TokenPair), args.Error(2)
}

func (m *MockAuthService) RefreshTokens(token string) (models.TokenPair, error) {
	args := m.Called(token)
	return args.Get(0).(models.TokenPair), args.Error(1)
}

func (m *MockAuthService) Logout(token string) error {
	return m.Called(token).Error(0)
}

func (m *MockAuthService) GetMe(userID uuid.UUID) (models.UserResponse, error) {
	args := m.Called(userID)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockAuthService) InvalidateAllSessions(userID uuid.UUID) error {
	return m.Called(userID).Error(0)
}

func (m *MockAuthService) RequestPasswordReset(email string) (string, error) {
	args := m.Called(email)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAuthService) ResetPassword(token, newPassword string) error {
	return m.Called(token, newPassword).Error(0)
}
