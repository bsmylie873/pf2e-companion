package mocks

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"pf2e-companion/backend/models"
)

type MockPartyMarkerRepository struct{ mock.Mock }

func (m *MockPartyMarkerRepository) FindByGameID(gameID uuid.UUID) (*models.PartyMarker, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PartyMarker), args.Error(1)
}

func (m *MockPartyMarkerRepository) Upsert(marker *models.PartyMarker) error {
	return m.Called(marker).Error(0)
}

func (m *MockPartyMarkerRepository) Delete(gameID uuid.UUID) error {
	return m.Called(gameID).Error(0)
}
