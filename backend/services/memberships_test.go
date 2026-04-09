package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestMembershipService_CreateMembership_Success(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := &models.GameMembership{GameID: gameID, UserID: uuid.New(), IsGM: false}
	callerMembership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(callerMembership, nil)
	mockMemberRepo.On("Create", membership).Return(nil)

	result, err := svc.CreateMembership(membership, callerID)

	assert.NoError(t, err)
	assert.Equal(t, membership.UserID, result.UserID)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_CreateMembership_Forbidden(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := &models.GameMembership{GameID: gameID, UserID: uuid.New()}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateMembership(membership, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_ListMemberships_Success(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	gameID := uuid.New()
	callerMembership := models.GameMembership{UserID: callerID, GameID: gameID}
	memberships := []models.GameMembership{callerMembership}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(callerMembership, nil)
	mockMemberRepo.On("FindByGameID", gameID).Return(memberships, nil)

	result, err := svc.ListMemberships(gameID, callerID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_ListMemberships_Forbidden(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(models.GameMembership{}, errors.New("not found"))

	_, err := svc.ListMemberships(gameID, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_GetMembership_Success(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	membershipID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{ID: membershipID, GameID: gameID, UserID: uuid.New()}
	callerMembership := models.GameMembership{UserID: callerID, GameID: gameID}

	mockMemberRepo.On("FindByID", membershipID).Return(membership, nil)
	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(callerMembership, nil)

	result, err := svc.GetMembership(membershipID, callerID)

	assert.NoError(t, err)
	assert.Equal(t, membershipID, result.ID)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_GetMembership_Forbidden(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	membershipID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{ID: membershipID, GameID: gameID, UserID: uuid.New()}

	mockMemberRepo.On("FindByID", membershipID).Return(membership, nil)
	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetMembership(membershipID, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_UpdateMembership_Success(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	membershipID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{ID: membershipID, GameID: gameID}
	callerMembership := models.GameMembership{UserID: callerID, GameID: gameID}
	updates := map[string]interface{}{"is_gm": true}
	updated := models.GameMembership{ID: membershipID, IsGM: true}

	mockMemberRepo.On("FindByID", membershipID).Return(membership, nil)
	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(callerMembership, nil)
	mockMemberRepo.On("Update", membershipID, updates).Return(updated, nil)

	result, err := svc.UpdateMembership(membershipID, callerID, updates)

	assert.NoError(t, err)
	assert.True(t, result.IsGM)
	mockMemberRepo.AssertExpectations(t)
}

func TestMembershipService_DeleteMembership_Success(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	membershipID := uuid.New()
	gameID := uuid.New()
	memberUserID := uuid.New()
	membership := models.GameMembership{ID: membershipID, GameID: gameID, UserID: memberUserID}
	callerMembership := models.GameMembership{UserID: callerID, GameID: gameID}

	mockMemberRepo.On("FindByID", membershipID).Return(membership, nil)
	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(callerMembership, nil)
	mockMemberRepo.On("Delete", membershipID).Return(nil)
	mockPrefSvc.On("ClearDefaultGameForMembership", memberUserID, gameID).Return(nil)

	err := svc.DeleteMembership(membershipID, callerID)

	assert.NoError(t, err)
	mockMemberRepo.AssertExpectations(t)
	mockPrefSvc.AssertExpectations(t)
}

func TestMembershipService_DeleteMembership_Forbidden(t *testing.T) {
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPrefSvc := &mocks.MockPreferenceService{}
	svc := NewMembershipService(mockMemberRepo, mockPrefSvc)

	callerID := uuid.New()
	membershipID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{ID: membershipID, GameID: gameID}

	mockMemberRepo.On("FindByID", membershipID).Return(membership, nil)
	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeleteMembership(membershipID, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}
