package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockInviteService struct{ mock.Mock }

func (m *MockInviteService) GenerateInvite(gameID, callerID uuid.UUID, expiry string) (models.InviteTokenResponse, error) {
	args := m.Called(gameID, callerID, expiry)
	return args.Get(0).(models.InviteTokenResponse), args.Error(1)
}

func (m *MockInviteService) GetActiveInvite(gameID, callerID uuid.UUID) (models.InviteTokenStatusResponse, error) {
	args := m.Called(gameID, callerID)
	return args.Get(0).(models.InviteTokenStatusResponse), args.Error(1)
}

func (m *MockInviteService) RevokeInvite(gameID, callerID uuid.UUID) error {
	return m.Called(gameID, callerID).Error(0)
}

func (m *MockInviteService) ValidateInvite(token string) (models.InviteValidationResponse, error) {
	args := m.Called(token)
	return args.Get(0).(models.InviteValidationResponse), args.Error(1)
}

func (m *MockInviteService) RedeemInvite(token string, userID uuid.UUID) (models.InviteRedeemResponse, error) {
	args := m.Called(token, userID)
	return args.Get(0).(models.InviteRedeemResponse), args.Error(1)
}
