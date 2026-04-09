package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockRefreshTokenRepository struct{ mock.Mock }

func (m *MockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	return m.Called(token).Error(0)
}

func (m *MockRefreshTokenRepository) FindByTokenHash(hash string) (models.RefreshToken, error) {
	args := m.Called(hash)
	return args.Get(0).(models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByTokenHash(hash string) error {
	return m.Called(hash).Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpiredForUser(userID uuid.UUID) error {
	return m.Called(userID).Error(0)
}

func (m *MockRefreshTokenRepository) DeleteAllForUser(userID uuid.UUID) error {
	return m.Called(userID).Error(0)
}
