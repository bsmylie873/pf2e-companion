package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPasswordResetTokenRepository struct{ mock.Mock }

func (m *MockPasswordResetTokenRepository) Create(token *models.PasswordResetToken) error {
	return m.Called(token).Error(0)
}

func (m *MockPasswordResetTokenRepository) FindByTokenHash(hash string) (models.PasswordResetToken, error) {
	args := m.Called(hash)
	return args.Get(0).(models.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetTokenRepository) MarkUsed(hash string) error {
	return m.Called(hash).Error(0)
}

func (m *MockPasswordResetTokenRepository) DeleteExpiredForUser(userID uuid.UUID) error {
	return m.Called(userID).Error(0)
}

func (m *MockPasswordResetTokenRepository) DeleteAllForUser(userID uuid.UUID) error {
	return m.Called(userID).Error(0)
}
