package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

// helper to create service + its two mock deps
func newPartyMarkerSvc() (PartyMarkerService, *mocks.MockPartyMarkerRepository, *mocks.MockMembershipRepository) {
	repoMock := &mocks.MockPartyMarkerRepository{}
	memberMock := &mocks.MockMembershipRepository{}
	svc := NewPartyMarkerService(repoMock, memberMock)
	return svc, repoMock, memberMock
}

// -- GetPartyMarker --

func TestPartyMarkerService_GetPartyMarker_Success(t *testing.T) {
	svc, repoMock, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()
	marker := &models.PartyMarker{ID: uuid.New(), GameID: gameID}

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, nil)
	repoMock.On("FindByGameID", gameID).Return(marker, nil)

	result, err := svc.GetPartyMarker(gameID, userID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, marker.ID, result.ID)
	memberMock.AssertExpectations(t)
	repoMock.AssertExpectations(t)
}

func TestPartyMarkerService_GetPartyMarker_Forbidden(t *testing.T) {
	svc, _, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	result, err := svc.GetPartyMarker(gameID, userID)
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, result)
	memberMock.AssertExpectations(t)
}

func TestPartyMarkerService_GetPartyMarker_ReturnsNilWhenNotFound(t *testing.T) {
	svc, repoMock, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, nil)
	repoMock.On("FindByGameID", gameID).Return((*models.PartyMarker)(nil), nil)

	result, err := svc.GetPartyMarker(gameID, userID)
	assert.NoError(t, err)
	assert.Nil(t, result)
	memberMock.AssertExpectations(t)
	repoMock.AssertExpectations(t)
}

// -- UpsertPartyMarker --

func TestPartyMarkerService_UpsertPartyMarker_Success(t *testing.T) {
	svc, repoMock, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, nil)
	repoMock.On("Upsert", mock.AnythingOfType("*models.PartyMarker")).Return(nil)

	result, err := svc.UpsertPartyMarker(gameID, userID, mapID, 0.5, 0.3)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, mapID, result.MapID)
	assert.Equal(t, 0.5, result.X)
	assert.Equal(t, 0.3, result.Y)
	memberMock.AssertExpectations(t)
	repoMock.AssertExpectations(t)
}

func TestPartyMarkerService_UpsertPartyMarker_Forbidden(t *testing.T) {
	svc, _, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, errors.New("not found"))

	result, err := svc.UpsertPartyMarker(gameID, userID, mapID, 0.5, 0.3)
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Nil(t, result)
	memberMock.AssertExpectations(t)
}

func TestPartyMarkerService_UpsertPartyMarker_InvalidX(t *testing.T) {
	svc, _, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, nil)

	result, err := svc.UpsertPartyMarker(gameID, userID, mapID, 150.0, 0.5)
	assert.ErrorIs(t, err, ErrValidation)
	assert.Nil(t, result)
	memberMock.AssertExpectations(t)
}

func TestPartyMarkerService_UpsertPartyMarker_InvalidY(t *testing.T) {
	svc, _, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, nil)

	result, err := svc.UpsertPartyMarker(gameID, userID, mapID, 0.5, -1.0)
	assert.ErrorIs(t, err, ErrValidation)
	assert.Nil(t, result)
	memberMock.AssertExpectations(t)
}

// -- DeletePartyMarker --

func TestPartyMarkerService_DeletePartyMarker_Success(t *testing.T) {
	svc, repoMock, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, nil)
	repoMock.On("Delete", gameID).Return(nil)

	err := svc.DeletePartyMarker(gameID, userID)
	assert.NoError(t, err)
	memberMock.AssertExpectations(t)
	repoMock.AssertExpectations(t)
}

func TestPartyMarkerService_DeletePartyMarker_Forbidden(t *testing.T) {
	svc, _, memberMock := newPartyMarkerSvc()

	userID := uuid.New()
	gameID := uuid.New()

	memberMock.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeletePartyMarker(gameID, userID)
	assert.ErrorIs(t, err, ErrForbidden)
	memberMock.AssertExpectations(t)
}
