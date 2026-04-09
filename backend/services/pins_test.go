package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestPinService_CreatePin_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	gameID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin := &models.SessionPin{Label: "Battle Site", X: 0.5, Y: 0.3}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Create", pin).Return(nil)

	result, err := svc.CreatePin(sessionID, userID, pin)

	assert.NoError(t, err)
	assert.Equal(t, "Battle Site", result.Label)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, &sessionID, result.SessionID)
	mockPinRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_CreatePin_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	gameID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreatePin(sessionID, userID, &models.SessionPin{Label: "X"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_CreatePin_SessionNotFound(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	sessionID := uuid.New()

	mockSessionRepo.On("FindByID", sessionID).Return(models.Session{}, gorm.ErrRecordNotFound)

	_, err := svc.CreatePin(sessionID, userID, &models.SessionPin{Label: "X"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockSessionRepo.AssertExpectations(t)
}

func TestPinService_CreateGamePin_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin := &models.SessionPin{Label: "Town"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Create", pin).Return(nil)

	result, err := svc.CreateGamePin(gameID, userID, pin)

	assert.NoError(t, err)
	assert.Equal(t, "Town", result.Label)
	assert.Equal(t, gameID, result.GameID)
	assert.Nil(t, result.SessionID)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_CreateGamePin_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateGamePin(gameID, userID, &models.SessionPin{Label: "X"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_ListGamePins_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pins := []models.SessionPin{{ID: uuid.New(), Label: "Pin 1"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByGameID", gameID).Return(pins, nil)

	result, err := svc.ListGamePins(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_ListGamePins_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListGamePins(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_GetPin_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID, Label: "Camp"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetPin(pinID, userID)

	assert.NoError(t, err)
	assert.Equal(t, pinID, result.ID)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_GetPin_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetPin(pinID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_UpdatePin_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"label": "Updated"}
	updated := models.SessionPin{ID: pinID, Label: "Updated"}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Update", pinID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	result, err := svc.UpdatePin(pinID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Updated", result.Label)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_UpdatePin_FindByIDError(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()

	mockPinRepo.On("FindByID", pinID).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdatePin(pinID, userID, map[string]interface{}{"label": "x"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinRepo.AssertExpectations(t)
}

func TestPinService_UpdatePin_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdatePin(pinID, userID, map[string]interface{}{"label": "x"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_UpdatePin_GroupedPinMove_Blocked_X(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID, GroupID: &groupID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"x": 0.5}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.UpdatePin(pinID, userID, updates)

	assert.ErrorIs(t, err, ErrGroupedPinMove)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_UpdatePin_GroupedPinMove_Blocked_Y(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID, GroupID: &groupID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"y": 0.7}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.UpdatePin(pinID, userID, updates)

	assert.ErrorIs(t, err, ErrGroupedPinMove)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_DeletePin_Success_NoGroup(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Delete", pinID).Return(nil)

	err := svc.DeletePin(pinID, userID)

	assert.NoError(t, err)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_DeletePin_FindByIDError(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()

	mockPinRepo.On("FindByID", pinID).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	err := svc.DeletePin(pinID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinRepo.AssertExpectations(t)
}

func TestPinService_DeletePin_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeletePin(pinID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_DeletePin_WithGroup_CountEquals0(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID, GroupID: &groupID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Delete", pinID).Return(nil)
	mockPinGroupRepo.On("CountMembers", groupID).Return(int64(0), nil)
	mockPinGroupRepo.On("Delete", groupID).Return(nil)

	err := svc.DeletePin(pinID, userID)

	assert.NoError(t, err)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinGroupRepo.AssertExpectations(t)
}

func TestPinService_DeletePin_WithGroup_CountEquals1_AutoDissolve(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	remainingPinID := uuid.New()
	pin := models.SessionPin{ID: pinID, GameID: gameID, GroupID: &groupID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	remaining := []models.SessionPin{{ID: remainingPinID}}

	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Delete", pinID).Return(nil)
	mockPinGroupRepo.On("CountMembers", groupID).Return(int64(1), nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(remaining, nil)
	mockPinRepo.On("ClearGroupID", remainingPinID).Return(nil)
	mockPinGroupRepo.On("Delete", groupID).Return(nil)

	err := svc.DeletePin(pinID, userID)

	assert.NoError(t, err)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinGroupRepo.AssertExpectations(t)
}

func TestPinService_CreateMapPin_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin := &models.SessionPin{Label: "Dungeon"}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("Create", pin).Return(nil)

	result, err := svc.CreateMapPin(mapID, userID, pin)

	assert.NoError(t, err)
	assert.Equal(t, "Dungeon", result.Label)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, &mapID, result.MapID)
	mockPinRepo.AssertExpectations(t)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_CreateMapPin_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateMapPin(mapID, userID, &models.SessionPin{Label: "X"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_ListMapPins_Success(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pins := []models.SessionPin{{ID: uuid.New(), Label: "Pin"}}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByMapID", mapID).Return(pins, nil)

	result, err := svc.ListMapPins(mapID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockPinRepo.AssertExpectations(t)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinService_ListMapPins_Forbidden(t *testing.T) {
	mockPinRepo := &mocks.MockPinRepository{}
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinService(mockPinRepo, mockSessionRepo, mockMemberRepo, mockPinGroupRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListMapPins(mapID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
