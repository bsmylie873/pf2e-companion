package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestInviteService_GenerateInvite_Success(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("RevokeAllForGame", gameID).Return(nil)
	mockInviteRepo.On("Create", mock.AnythingOfType("*models.InviteToken")).Return(nil)

	result, err := svc.GenerateInvite(gameID, callerID, "24h")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.NotNil(t, result.ExpiresAt)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GenerateInvite_NotGM(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)

	_, err := svc.GenerateInvite(gameID, callerID, "24h")

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GenerateInvite_InvalidExpiresIn(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)

	_, err := svc.GenerateInvite(gameID, callerID, "30m")

	assert.ErrorIs(t, err, ErrValidation)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GenerateInvite_CreateError(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}
	createErr := errors.New("db insert error")

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("RevokeAllForGame", gameID).Return(nil)
	mockInviteRepo.On("Create", mock.AnythingOfType("*models.InviteToken")).Return(createErr)

	_, err := svc.GenerateInvite(gameID, callerID, "24h")

	assert.ErrorIs(t, err, createErr)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GenerateInvite_7d(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("RevokeAllForGame", gameID).Return(nil)
	mockInviteRepo.On("Create", mock.AnythingOfType("*models.InviteToken")).Return(nil)

	result, err := svc.GenerateInvite(gameID, callerID, "7d")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.NotNil(t, result.ExpiresAt)
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_GenerateInvite_Never(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("RevokeAllForGame", gameID).Return(nil)
	mockInviteRepo.On("Create", mock.AnythingOfType("*models.InviteToken")).Return(nil)

	result, err := svc.GenerateInvite(gameID, callerID, "never")

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Nil(t, result.ExpiresAt)
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_GetActiveInvite_Success(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}
	token := models.InviteToken{GameID: gameID}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("FindActiveByGameID", gameID).Return(token, nil)

	result, err := svc.GetActiveInvite(gameID, callerID)

	assert.NoError(t, err)
	assert.True(t, result.HasActiveInvite)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GetActiveInvite_NoActiveInvite(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("FindActiveByGameID", gameID).Return(models.InviteToken{}, gorm.ErrRecordNotFound)

	result, err := svc.GetActiveInvite(gameID, callerID)

	assert.NoError(t, err)
	assert.False(t, result.HasActiveInvite)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GetActiveInvite_Forbidden(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)

	_, err := svc.GetActiveInvite(gameID, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_GetActiveInvite_DBError(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}
	dbErr := errors.New("connection error")

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("FindActiveByGameID", gameID).Return(models.InviteToken{}, dbErr)

	_, err := svc.GetActiveInvite(gameID, callerID)

	assert.ErrorIs(t, err, dbErr)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_RevokeInvite_Success(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)
	mockInviteRepo.On("RevokeAllForGame", gameID).Return(nil)

	err := svc.RevokeInvite(gameID, callerID)

	assert.NoError(t, err)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_RevokeInvite_Forbidden(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	callerID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: callerID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", callerID, gameID).Return(membership, nil)

	err := svc.RevokeInvite(gameID, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_ValidateInvite_Success(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "valid-raw-token-string-abc123"
	hash := hashInviteToken(rawToken)
	gameID := uuid.New()
	token := models.InviteToken{GameID: gameID}
	game := models.Game{ID: gameID, Title: "My Game"}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)
	mockGameRepo.On("FindByID", gameID).Return(game, nil)

	result, err := svc.ValidateInvite(rawToken)

	assert.NoError(t, err)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, "My Game", result.GameTitle)
	mockInviteRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestInviteService_ValidateInvite_NotFound(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "missing-token"
	hash := hashInviteToken(rawToken)

	mockInviteRepo.On("FindByTokenHash", hash).Return(models.InviteToken{}, gorm.ErrRecordNotFound)

	_, err := svc.ValidateInvite(rawToken)

	assert.EqualError(t, err, "invalid invite token")
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_ValidateInvite_Revoked(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "revoked-token-abc"
	hash := hashInviteToken(rawToken)
	revokedAt := time.Now().Add(-1 * time.Hour)
	token := models.InviteToken{RevokedAt: &revokedAt}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)

	_, err := svc.ValidateInvite(rawToken)

	assert.EqualError(t, err, "invite token has been revoked")
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_ValidateInvite_Expired(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "expired-token-abc"
	hash := hashInviteToken(rawToken)
	expired := time.Now().Add(-1 * time.Hour)
	token := models.InviteToken{ExpiresAt: &expired}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)

	_, err := svc.ValidateInvite(rawToken)

	assert.EqualError(t, err, "invite token has expired")
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_ValidateInvite_GameNotFound(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "valid-token-no-game"
	hash := hashInviteToken(rawToken)
	gameID := uuid.New()
	token := models.InviteToken{GameID: gameID}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)
	mockGameRepo.On("FindByID", gameID).Return(models.Game{}, gorm.ErrRecordNotFound)

	_, err := svc.ValidateInvite(rawToken)

	assert.EqualError(t, err, "game not found")
	mockInviteRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestInviteService_RedeemInvite_Success_NewMember(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "redeem-token-xyz"
	hash := hashInviteToken(rawToken)
	gameID := uuid.New()
	userID := uuid.New()
	token := models.InviteToken{GameID: gameID}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)
	mockMemberRepo.On("Create", mock.AnythingOfType("*models.GameMembership")).Return(nil)

	result, err := svc.RedeemInvite(rawToken, userID)

	assert.NoError(t, err)
	assert.Equal(t, gameID, result.GameID)
	assert.False(t, result.AlreadyMember)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_RedeemInvite_AlreadyMember(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "redeem-token-existing"
	hash := hashInviteToken(rawToken)
	gameID := uuid.New()
	userID := uuid.New()
	membershipID := uuid.New()
	token := models.InviteToken{GameID: gameID}
	existing := models.GameMembership{ID: membershipID, GameID: gameID, UserID: userID}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(existing, nil)

	result, err := svc.RedeemInvite(rawToken, userID)

	assert.NoError(t, err)
	assert.True(t, result.AlreadyMember)
	assert.Equal(t, membershipID, result.MembershipID)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestInviteService_RedeemInvite_NotFound(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "missing-redeem-token"
	hash := hashInviteToken(rawToken)

	mockInviteRepo.On("FindByTokenHash", hash).Return(models.InviteToken{}, gorm.ErrRecordNotFound)

	_, err := svc.RedeemInvite(rawToken, uuid.New())

	assert.EqualError(t, err, "invalid invite token")
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_RedeemInvite_Revoked(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "revoked-redeem-token"
	hash := hashInviteToken(rawToken)
	revokedAt := time.Now().Add(-1 * time.Hour)
	token := models.InviteToken{RevokedAt: &revokedAt}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)

	_, err := svc.RedeemInvite(rawToken, uuid.New())

	assert.EqualError(t, err, "invite token has been revoked")
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_RedeemInvite_Expired(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "expired-redeem-token"
	hash := hashInviteToken(rawToken)
	expired := time.Now().Add(-1 * time.Hour)
	token := models.InviteToken{ExpiresAt: &expired}

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)

	_, err := svc.RedeemInvite(rawToken, uuid.New())

	assert.EqualError(t, err, "invite token has expired")
	mockInviteRepo.AssertExpectations(t)
}

func TestInviteService_RedeemInvite_CreateError(t *testing.T) {
	mockInviteRepo := &mocks.MockInviteTokenRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockGameRepo := &mocks.MockGameRepository{}
	svc := NewInviteService(mockInviteRepo, mockMemberRepo, mockGameRepo)

	rawToken := "valid-redeem-token"
	hash := hashInviteToken(rawToken)
	gameID := uuid.New()
	userID := uuid.New()
	token := models.InviteToken{GameID: gameID}
	createErr := errors.New("membership create error")

	mockInviteRepo.On("FindByTokenHash", hash).Return(token, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)
	mockMemberRepo.On("Create", mock.AnythingOfType("*models.GameMembership")).Return(createErr)

	_, err := svc.RedeemInvite(rawToken, userID)

	assert.ErrorIs(t, err, createErr)
	mockInviteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
