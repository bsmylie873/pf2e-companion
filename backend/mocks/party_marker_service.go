package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPartyMarkerService struct{ mock.Mock }

func (m *MockPartyMarkerService) GetPartyMarker(gameID, userID uuid.UUID) (*models.PartyMarker, error) {
	args := m.Called(gameID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PartyMarker), args.Error(1)
}

func (m *MockPartyMarkerService) UpsertPartyMarker(gameID, userID uuid.UUID, mapID uuid.UUID, x, y float64) (*models.PartyMarker, error) {
	args := m.Called(gameID, userID, mapID, x, y)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PartyMarker), args.Error(1)
}

func (m *MockPartyMarkerService) DeletePartyMarker(gameID, userID uuid.UUID) error {
	return m.Called(gameID, userID).Error(0)
}
