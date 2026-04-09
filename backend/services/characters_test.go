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

func TestCharacterService_CreateCharacter_Success(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	char := &models.Character{Name: "Aragorn"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockCharRepo.On("Create", char).Return(nil)

	result, err := svc.CreateCharacter(gameID, userID, char)

	assert.NoError(t, err)
	assert.Equal(t, "Aragorn", result.Name)
	assert.Equal(t, gameID, result.GameID)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_CreateCharacter_Forbidden(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateCharacter(gameID, userID, &models.Character{Name: "X"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_ListGameCharacters_Success(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	chars := []models.Character{{ID: uuid.New(), Name: "Aragorn"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockCharRepo.On("FindByGameID", gameID).Return(chars, nil)

	result, err := svc.ListGameCharacters(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_ListGameCharacters_Forbidden(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListGameCharacters(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_GetCharacter_Success(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID, Name: "Legolas"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetCharacter(charID, userID)

	assert.NoError(t, err)
	assert.Equal(t, charID, result.ID)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_GetCharacter_Forbidden(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetCharacter(charID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_UpdateCharacter_Success(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"name": "Gimli"}
	updated := models.Character{ID: charID, Name: "Gimli"}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockCharRepo.On("Update", charID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	result, err := svc.UpdateCharacter(charID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Gimli", result.Name)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_DeleteCharacter_Success(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockCharRepo.On("Delete", charID).Return(nil)

	err := svc.DeleteCharacter(charID, userID)

	assert.NoError(t, err)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestCharacterService_DeleteCharacter_Forbidden(t *testing.T) {
	mockCharRepo := &mocks.MockCharacterRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewCharacterService(mockCharRepo, mockMemberRepo)

	userID := uuid.New()
	charID := uuid.New()
	gameID := uuid.New()
	char := models.Character{ID: charID, GameID: gameID}

	mockCharRepo.On("FindByID", charID).Return(char, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeleteCharacter(charID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockCharRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
