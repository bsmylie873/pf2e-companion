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

func TestGameService_CreateGame_NoMembers(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	game := &models.Game{Title: "Test Game"}
	creatorID := uuid.New()

	mockGameRepo.On("Create", game).Return(nil)
	mockMemberRepo.On("Create", mock.MatchedBy(func(m *models.GameMembership) bool {
		return m.UserID == creatorID && m.IsGM == true
	})).Return(nil)

	result, err := svc.CreateGame(game, nil, creatorID)

	assert.NoError(t, err)
	assert.Equal(t, "Test Game", result.Title)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_CreateGame_WithMembers(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	gameID := uuid.New()
	game := &models.Game{ID: gameID, Title: "Test Game"}
	creatorID := uuid.New()
	members := []models.GameMembership{
		{UserID: uuid.New(), IsGM: false},
	}

	mockGameRepo.On("Create", game).Return(nil)
	mockMemberRepo.On("Create", mock.AnythingOfType("*models.GameMembership")).Return(nil)

	result, err := svc.CreateGame(game, members, creatorID)

	assert.NoError(t, err)
	assert.Equal(t, "Test Game", result.Title)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_CreateGame_RepoError(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	game := &models.Game{Title: "Test Game"}
	repoErr := errors.New("db error")

	mockGameRepo.On("Create", game).Return(repoErr)

	_, err := svc.CreateGame(game, nil, uuid.New())

	assert.ErrorIs(t, err, repoErr)
	mockGameRepo.AssertExpectations(t)
}

func TestGameService_ListGames_Success(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	memberships := []models.GameMembership{{GameID: gameID, UserID: userID, IsGM: true}}
	repoGames := []models.Game{{ID: gameID, Title: "Game 1"}}
	expectedGames := []models.GameWithRole{{Game: models.Game{ID: gameID, Title: "Game 1"}, IsGM: true}}

	mockMemberRepo.On("FindByUserID", userID).Return(memberships, nil)
	mockGameRepo.On("FindByIDs", []uuid.UUID{gameID}).Return(repoGames, nil)

	games, err := svc.ListGames(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGames, games)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_ListGames_MembershipError(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	repoErr := errors.New("db error")

	mockMemberRepo.On("FindByUserID", userID).Return([]models.GameMembership{}, repoErr)

	_, err := svc.ListGames(userID)

	assert.Error(t, err)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_ListGamesPaginated_Success(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	memberships := []models.GameMembership{{GameID: gameID, UserID: userID, IsGM: true}}
	repoGames := []models.Game{{ID: gameID, Title: "Game 1"}}
	expectedGames := []models.GameWithRole{{Game: models.Game{ID: gameID, Title: "Game 1"}, IsGM: true}}
	var total int64 = 1

	mockMemberRepo.On("FindByUserID", userID).Return(memberships, nil)
	mockGameRepo.On("FindByIDsPaginated", []uuid.UUID{gameID}, 0, 10).Return(repoGames, total, nil)

	games, count, err := svc.ListGamesPaginated(userID, 0, 10)

	assert.NoError(t, err)
	assert.Equal(t, expectedGames, games)
	assert.Equal(t, int64(1), count)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_GetGame_Success(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	expectedGame := models.Game{ID: gameID, Title: "Test Game"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockGameRepo.On("FindByID", gameID).Return(expectedGame, nil)

	game, err := svc.GetGame(gameID, userID)

	assert.NoError(t, err)
	assert.Equal(t, gameID, game.ID)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_GetGame_Forbidden(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetGame(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_UpdateGame_Success(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	updates := map[string]interface{}{"title": "Updated", "id": "ignored", "created_at": "ignored"}
	cleanUpdates := map[string]interface{}{"title": "Updated"}
	expectedGame := models.Game{ID: gameID, Title: "Updated"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockGameRepo.On("Update", gameID, cleanUpdates).Return(expectedGame, nil)

	game, err := svc.UpdateGame(gameID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Updated", game.Title)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_UpdateGame_Forbidden(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdateGame(gameID, userID, map[string]interface{}{"title": "X"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_DeleteGame_Success(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockGameRepo.On("Delete", gameID).Return(nil)

	err := svc.DeleteGame(gameID, userID)

	assert.NoError(t, err)
	mockGameRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_DeleteGame_Forbidden(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeleteGame(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestGameService_UpdateGame_NonGM_Forbidden(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.UpdateGame(gameID, userID, map[string]interface{}{})
	assert.ErrorIs(t, err, ErrForbidden)
	mockGameRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

func TestGameService_DeleteGame_NonGM_Forbidden(t *testing.T) {
	mockGameRepo := &mocks.MockGameRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewGameService(mockGameRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	err := svc.DeleteGame(gameID, userID)
	assert.ErrorIs(t, err, ErrForbidden)
	mockGameRepo.AssertNotCalled(t, "Delete", mock.Anything)
}
