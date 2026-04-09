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

func TestItemService_CreateItem_Success(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	item := &models.Item{Name: "Sword"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockItemRepo.On("Create", item).Return(nil)

	result, err := svc.CreateItem(gameID, userID, item)

	assert.NoError(t, err)
	assert.Equal(t, "Sword", result.Name)
	assert.Equal(t, gameID, result.GameID)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_CreateItem_Forbidden(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateItem(gameID, userID, &models.Item{Name: "X"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_ListGameItems_Success(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	items := []models.Item{{ID: uuid.New(), Name: "Sword"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockItemRepo.On("FindByGameID", gameID).Return(items, nil)

	result, err := svc.ListGameItems(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_ListGameItems_Forbidden(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListGameItems(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_ListCharacterItems_Success(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	items := []models.Item{{ID: uuid.New(), Name: "Shield"}}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockItemRepo.On("FindByCharacterID", charID).Return(items, nil)

	result, err := svc.ListCharacterItems(charID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockCharRepo.AssertExpectations(t)
}

func TestItemService_ListCharacterItems_Forbidden(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListCharacterItems(charID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_GetItem_Success(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	itemID := uuid.New()
	gameID := uuid.New()
	item := models.Item{ID: itemID, GameID: gameID, Name: "Sword"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockItemRepo.On("FindByID", itemID).Return(item, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetItem(itemID, userID)

	assert.NoError(t, err)
	assert.Equal(t, itemID, result.ID)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_GetItem_Forbidden(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	itemID := uuid.New()
	gameID := uuid.New()
	item := models.Item{ID: itemID, GameID: gameID}

	mockItemRepo.On("FindByID", itemID).Return(item, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetItem(itemID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_UpdateItem_Success(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	itemID := uuid.New()
	gameID := uuid.New()
	item := models.Item{ID: itemID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"name": "Axe"}
	updated := models.Item{ID: itemID, Name: "Axe"}

	mockItemRepo.On("FindByID", itemID).Return(item, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockItemRepo.On("Update", itemID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	result, err := svc.UpdateItem(itemID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Axe", result.Name)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_DeleteItem_Success(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	itemID := uuid.New()
	gameID := uuid.New()
	item := models.Item{ID: itemID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockItemRepo.On("FindByID", itemID).Return(item, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockItemRepo.On("Delete", itemID).Return(nil)

	err := svc.DeleteItem(itemID, userID)

	assert.NoError(t, err)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestItemService_DeleteItem_Forbidden(t *testing.T) {
	mockItemRepo := &mocks.MockItemRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockCharRepo := &mocks.MockCharacterRepository{}
	svc := NewItemService(mockItemRepo, mockMemberRepo, mockCharRepo)

	userID := uuid.New()
	itemID := uuid.New()
	gameID := uuid.New()
	item := models.Item{ID: itemID, GameID: gameID}

	mockItemRepo.On("FindByID", itemID).Return(item, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeleteItem(itemID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockItemRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
