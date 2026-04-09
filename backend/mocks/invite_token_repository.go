package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockInviteTokenRepository struct{ mock.Mock }

func (m *MockInviteTokenRepository) Create(token *models.InviteToken) error {
	return m.Called(token).Error(0)
}

func (m *MockInviteTokenRepository) FindActiveByGameID(gameID uuid.UUID) (models.InviteToken, error) {
	args := m.Called(gameID)
	return args.Get(0).(models.InviteToken), args.Error(1)
}

func (m *MockInviteTokenRepository) FindByTokenHash(hash string) (models.InviteToken, error) {
	args := m.Called(hash)
	return args.Get(0).(models.InviteToken), args.Error(1)
}

func (m *MockInviteTokenRepository) RevokeAllForGame(gameID uuid.UUID) error {
	return m.Called(gameID).Error(0)
}

func (m *MockInviteTokenRepository) RevokeByID(id uuid.UUID) error {
	return m.Called(id).Error(0)
}
