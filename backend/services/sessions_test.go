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

func TestSessionService_CreateSession_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	session := &models.Session{Title: "Session 1"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("Create", session).Return(nil)

	result, err := svc.CreateSession(gameID, userID, session)

	assert.NoError(t, err)
	assert.Equal(t, "Session 1", result.Title)
	assert.Equal(t, gameID, result.GameID)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_CreateSession_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateSession(gameID, userID, &models.Session{Title: "S"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_ListGameSessions_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	sessions := []models.Session{{ID: uuid.New(), Title: "S1", GameID: gameID}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("FindByGameID", gameID).Return(sessions, nil)

	result, err := svc.ListGameSessions(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_ListGameSessions_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListGameSessions(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_ListGameSessionsPaginated_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	sessions := []models.Session{{ID: uuid.New(), Title: "S1"}}
	var total int64 = 1

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("FindByGameIDPaginated", gameID, 0, 10).Return(sessions, total, nil)

	result, count, err := svc.ListGameSessionsPaginated(gameID, userID, 0, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), count)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_ListGameSessionsPaginated_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, _, err := svc.ListGameSessionsPaginated(gameID, userID, 0, 10)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_GetSession_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	sessionID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID, Title: "S1"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetSession(sessionID, userID)

	assert.NoError(t, err)
	assert.Equal(t, sessionID, result.ID)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_GetSession_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	sessionID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetSession(sessionID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_UpdateSession_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	sessionID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updates := map[string]interface{}{"title": "Updated"}
	updatedSession := models.Session{ID: sessionID, Title: "Updated"}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("Update", sessionID, mock.AnythingOfType("map[string]interface {}")).Return(updatedSession, nil)

	result, err := svc.UpdateSession(sessionID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Updated", result.Title)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_DeleteSession_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	sessionID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("Delete", sessionID).Return(nil)

	err := svc.DeleteSession(sessionID, userID)

	assert.NoError(t, err)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestSessionService_DeleteSession_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewSessionService(mockSessionRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	sessionID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeleteSession(sessionID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
