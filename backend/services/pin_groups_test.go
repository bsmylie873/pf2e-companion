package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestPinGroupService_CreateGroup_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	pinID1 := uuid.New()
	pinID2 := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin1 := models.SessionPin{ID: pinID1, GameID: gameID, X: 0.5, Y: 0.3, Colour: "red", Icon: "castle"}
	pin2 := models.SessionPin{ID: pinID2, GameID: gameID}
	groupID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	pins := []models.SessionPin{pin1, pin2}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID1).Return(pin1, nil)
	mockPinRepo.On("FindByID", pinID2).Return(pin2, nil)
	mockPinGroupRepo.On("Create", mock.AnythingOfType("*models.PinGroup")).Return(nil).Run(func(args mock.Arguments) {
		g := args.Get(0).(*models.PinGroup)
		g.ID = groupID
	})
	mockPinRepo.On("SetGroupID", pinID1, mock.AnythingOfType("uuid.UUID")).Return(nil)
	mockPinRepo.On("SetGroupID", pinID2, mock.AnythingOfType("uuid.UUID")).Return(nil)
	mockPinRepo.On("FindByGroupID", mock.AnythingOfType("uuid.UUID")).Return(pins, nil)

	_ = group
	result, err := svc.CreateGroup(gameID, userID, []uuid.UUID{pinID1, pinID2})

	assert.NoError(t, err)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, 2, result.PinCount)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateGroup_TooFewPins(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateGroup(gameID, userID, []uuid.UUID{uuid.New()})

	assert.EqualError(t, err, "at least 2 pins required to create a group")
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateGroup(gameID, userID, []uuid.UUID{uuid.New(), uuid.New()})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateGroup_PinNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	pinID1 := uuid.New()
	pinID2 := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID1).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateGroup(gameID, userID, []uuid.UUID{pinID1, pinID2})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateGroup_PinWrongGame(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	otherGameID := uuid.New()
	pinID1 := uuid.New()
	pinID2 := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin1 := models.SessionPin{ID: pinID1, GameID: otherGameID} // wrong game

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID1).Return(pin1, nil)

	_, err := svc.CreateGroup(gameID, userID, []uuid.UUID{pinID1, pinID2})

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateGroup_PinAlreadyInGroup(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	pinID1 := uuid.New()
	pinID2 := uuid.New()
	existingGroupID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin1 := models.SessionPin{ID: pinID1, GameID: gameID, GroupID: &existingGroupID}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID1).Return(pin1, nil)

	_, err := svc.CreateGroup(gameID, userID, []uuid.UUID{pinID1, pinID2})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in a group")
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_GetGroup_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pins := []models.SessionPin{{ID: uuid.New()}}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pins, nil)

	result, err := svc.GetGroup(groupID, userID)

	assert.NoError(t, err)
	assert.Equal(t, groupID, result.ID)
	assert.Equal(t, 1, result.PinCount)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_GetGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetGroup(groupID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_AddPinToGroup_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	pin := models.SessionPin{ID: pinID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pinsAfter := []models.SessionPin{{ID: uuid.New()}, pin}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID).Return(pin, nil)
	mockPinRepo.On("SetGroupID", pinID, groupID).Return(nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pinsAfter, nil)

	result, err := svc.AddPinToGroup(groupID, pinID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.PinCount)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_AddPinToGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.AddPinToGroup(groupID, pinID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_AddPinToGroup_PinNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	_, err := svc.AddPinToGroup(groupID, pinID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_AddPinToGroup_PinWrongGame(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	otherGameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	pin := models.SessionPin{ID: pinID, GameID: otherGameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID).Return(pin, nil)

	_, err := svc.AddPinToGroup(groupID, pinID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_AddPinToGroup_PinAlreadyInGroup(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	existingGroup := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	pin := models.SessionPin{ID: pinID, GameID: gameID, GroupID: &existingGroup}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID).Return(pin, nil)

	_, err := svc.AddPinToGroup(groupID, pinID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in a group")
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_RemovePinFromGroup_CountGreaterThan1(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	remainingPins := []models.SessionPin{{ID: uuid.New()}, {ID: uuid.New()}}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("ClearGroupID", pinID).Return(nil)
	mockPinGroupRepo.On("CountMembers", groupID).Return(int64(2), nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(remainingPins, nil)

	result, err := svc.RemovePinFromGroup(groupID, pinID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.PinCount)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_RemovePinFromGroup_CountEquals1_AutoDissolve(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	remainingPinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	remaining := []models.SessionPin{{ID: remainingPinID}}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("ClearGroupID", pinID).Return(nil)
	mockPinGroupRepo.On("CountMembers", groupID).Return(int64(1), nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(remaining, nil)
	mockPinRepo.On("ClearGroupID", remainingPinID).Return(nil)
	mockPinGroupRepo.On("Delete", groupID).Return(nil)

	result, err := svc.RemovePinFromGroup(groupID, pinID, userID)

	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, result.ID) // returns empty response after dissolve
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_RemovePinFromGroup_CountEquals0_DeleteGroup(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("ClearGroupID", pinID).Return(nil)
	mockPinGroupRepo.On("CountMembers", groupID).Return(int64(0), nil)
	mockPinGroupRepo.On("Delete", groupID).Return(nil)

	result, err := svc.RemovePinFromGroup(groupID, pinID, userID)

	assert.NoError(t, err)
	assert.Equal(t, uuid.Nil, result.ID)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_UpdateGroup_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"colour": "blue"}
	updatedGroup := models.PinGroup{ID: groupID, Colour: "blue"}
	pins := []models.SessionPin{}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinGroupRepo.On("Update", groupID, updates).Return(updatedGroup, nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pins, nil)

	result, err := svc.UpdateGroup(groupID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "blue", result.Colour)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_DisbandGroup_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pins := []models.SessionPin{{ID: uuid.New()}, {ID: uuid.New()}}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pins, nil)
	mockPinRepo.On("ClearGroupID", mock.AnythingOfType("uuid.UUID")).Return(nil)
	mockPinGroupRepo.On("Delete", groupID).Return(nil)

	err := svc.DisbandGroup(groupID, userID)

	assert.NoError(t, err)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_ListGameGroups_Success_GM(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	groups := []models.PinGroup{{ID: groupID, GameID: gameID}}
	pins := []models.SessionPin{{ID: uuid.New()}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinGroupRepo.On("FindByGameID", gameID).Return(groups, nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pins, nil)

	result, err := svc.ListGameGroups(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].PinCount)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_ListGameGroups_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListGameGroups(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_ListGameGroups_NonGM_FiltersPrivateNote(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	ownerID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	noteID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	groups := []models.PinGroup{{ID: groupID, GameID: gameID}}
	// Pin linked to a private note owned by someone else
	pins := []models.SessionPin{{ID: uuid.New(), NoteID: &noteID}}
	privateNote := models.Note{ID: noteID, Visibility: "private", UserID: ownerID}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinGroupRepo.On("FindByGameID", gameID).Return(groups, nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pins, nil)
	mockNoteRepo.On("FindByID", noteID).Return(privateNote, nil)

	result, err := svc.ListGameGroups(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 0, result[0].PinCount) // private note pin filtered out
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateMapGroup_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	pinID1 := uuid.New()
	pinID2 := uuid.New()
	groupID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin1 := models.SessionPin{ID: pinID1, GameID: gameID, X: 0.1, Y: 0.2}
	pin2 := models.SessionPin{ID: pinID2, GameID: gameID}
	pins := []models.SessionPin{pin1, pin2}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID1).Return(pin1, nil)
	mockPinRepo.On("FindByID", pinID2).Return(pin2, nil)
	mockPinGroupRepo.On("Create", mock.AnythingOfType("*models.PinGroup")).Return(nil).Run(func(args mock.Arguments) {
		g := args.Get(0).(*models.PinGroup)
		g.ID = groupID
	})
	mockPinRepo.On("SetGroupID", pinID1, mock.AnythingOfType("uuid.UUID")).Return(nil)
	mockPinRepo.On("SetGroupID", pinID2, mock.AnythingOfType("uuid.UUID")).Return(nil)
	mockPinRepo.On("FindByGroupID", mock.AnythingOfType("uuid.UUID")).Return(pins, nil)

	result, err := svc.CreateMapGroup(mapID, userID, []uuid.UUID{pinID1, pinID2})

	assert.NoError(t, err)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, 2, result.PinCount)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateMapGroup_MapNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateMapGroup(mapID, userID, []uuid.UUID{uuid.New(), uuid.New()})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockMapRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateMapGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateMapGroup(mapID, userID, []uuid.UUID{uuid.New(), uuid.New()})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_ListMapGroups_Success(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	groupID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	groups := []models.PinGroup{{ID: groupID, GameID: gameID}}
	pins := []models.SessionPin{{ID: uuid.New()}}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinGroupRepo.On("FindByMapID", mapID).Return(groups, nil)
	mockPinRepo.On("FindByGroupID", groupID).Return(pins, nil)

	result, err := svc.ListMapGroups(mapID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].PinCount)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinGroupRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_ListMapGroups_MapNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	_, err := svc.ListMapGroups(mapID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockMapRepo.AssertExpectations(t)
}

func TestPinGroupService_ListMapGroups_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListMapGroups(mapID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateMapGroup_TooFewPins(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateMapGroup(mapID, userID, []uuid.UUID{uuid.New()})

	assert.EqualError(t, err, "at least 2 pins required to create a group")
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_CreateMapGroup_PinWrongGame(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	otherGameID := uuid.New()
	pinID1 := uuid.New()
	pinID2 := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	pin1 := models.SessionPin{ID: pinID1, GameID: otherGameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinRepo.On("FindByID", pinID1).Return(pin1, nil)

	_, err := svc.CreateMapGroup(mapID, userID, []uuid.UUID{pinID1, pinID2})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockPinRepo.AssertExpectations(t)
}

func TestPinGroupService_DisbandGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DisbandGroup(groupID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_DisbandGroup_GroupNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()

	mockPinGroupRepo.On("FindByID", groupID).Return(models.PinGroup{}, gorm.ErrRecordNotFound)

	err := svc.DisbandGroup(groupID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinGroupRepo.AssertExpectations(t)
}

func TestPinGroupService_UpdateGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdateGroup(groupID, userID, map[string]interface{}{"colour": "blue"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_UpdateGroup_GroupNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()

	mockPinGroupRepo.On("FindByID", groupID).Return(models.PinGroup{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdateGroup(groupID, userID, map[string]interface{}{"colour": "blue"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinGroupRepo.AssertExpectations(t)
}

func TestPinGroupService_UpdateGroup_UpdateError(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"colour": "blue"}
	updateErr := errors.New("update failed")

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinGroupRepo.On("Update", groupID, updates).Return(models.PinGroup{}, updateErr)

	_, err := svc.UpdateGroup(groupID, userID, updates)

	assert.ErrorIs(t, err, updateErr)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_RemovePinFromGroup_Forbidden(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()
	group := models.PinGroup{ID: groupID, GameID: gameID}

	mockPinGroupRepo.On("FindByID", groupID).Return(group, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.RemovePinFromGroup(groupID, pinID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPinGroupService_RemovePinFromGroup_GroupNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	mockPinGroupRepo.On("FindByID", groupID).Return(models.PinGroup{}, gorm.ErrRecordNotFound)

	_, err := svc.RemovePinFromGroup(groupID, pinID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinGroupRepo.AssertExpectations(t)
}

func TestPinGroupService_GetGroup_GroupNotFound(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	groupID := uuid.New()

	mockPinGroupRepo.On("FindByID", groupID).Return(models.PinGroup{}, gorm.ErrRecordNotFound)

	_, err := svc.GetGroup(groupID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockPinGroupRepo.AssertExpectations(t)
}

func TestPinGroupService_ListGameGroups_FindGroupsError(t *testing.T) {
	mockPinGroupRepo := &mocks.MockPinGroupRepository{}
	mockPinRepo := &mocks.MockPinRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockMapRepo := &mocks.MockMapRepository{}
	svc := NewPinGroupService(mockPinGroupRepo, mockPinRepo, mockNoteRepo, mockMemberRepo, mockMapRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	repoErr := errors.New("db error")

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPinGroupRepo.On("FindByGameID", gameID).Return([]models.PinGroup{}, repoErr)

	_, err := svc.ListGameGroups(gameID, userID)

	assert.ErrorIs(t, err, repoErr)
	mockPinGroupRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
