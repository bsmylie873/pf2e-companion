package services

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pf2e-companion/backend/auth"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func init() {
	os.Setenv("JWT_SECRET", "test-jwt-secret-for-auth-service-tests")
}

func TestAuthService_Register_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	req := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	mockTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockTokenRepo.On("DeleteExpiredForUser", mock.AnythingOfType("uuid.UUID")).Return(nil)

	userResp, pair, err := svc.Register(req)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", userResp.Username)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Register_CreateError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	req := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	dbErr := errors.New("duplicate key")

	mockUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(dbErr)

	_, _, err := svc.Register(req)

	assert.Error(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_TokenCreateError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	req := models.RegisterRequest{
		Username: "tokenErrUser",
		Email:    "tokenerr@example.com",
		Password: "password123",
	}
	tokenErr := errors.New("token store error")

	// User create succeeds, but token (issueTokenPair) create fails
	mockUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	mockTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(tokenErr)

	_, _, err := svc.Register(req)

	assert.ErrorIs(t, err, tokenErr)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	userID := uuid.New()
	user := models.User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashed),
	}

	mockUserRepo.On("FindByUsername", "testuser").Return(user, nil)
	mockTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockTokenRepo.On("DeleteExpiredForUser", userID).Return(nil)

	userResp, pair, err := svc.Login(models.LoginRequest{Username: "testuser", Password: "password123"})

	assert.NoError(t, err)
	assert.Equal(t, "testuser", userResp.Username)
	assert.NotEmpty(t, pair.AccessToken)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	mockUserRepo.On("FindByUsername", "unknown").Return(models.User{}, gorm.ErrRecordNotFound)

	_, _, err := svc.Login(models.LoginRequest{Username: "unknown", Password: "x"})

	assert.EqualError(t, err, "invalid credentials")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	user := models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: string(hashed),
	}

	mockUserRepo.On("FindByUsername", "testuser").Return(user, nil)

	_, _, err := svc.Login(models.LoginRequest{Username: "testuser", Password: "wrongpassword"})

	assert.EqualError(t, err, "invalid credentials")
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_RefreshTokens_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	refreshToken, _ := auth.GenerateRefreshToken(userID)
	tokenHash := hashToken(refreshToken)

	record := models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mockTokenRepo.On("FindByTokenHash", tokenHash).Return(record, nil)
	mockTokenRepo.On("DeleteByTokenHash", tokenHash).Return(nil)
	mockTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockTokenRepo.On("DeleteExpiredForUser", userID).Return(nil)

	pair, err := svc.RefreshTokens(refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_RefreshTokens_InvalidToken(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	_, err := svc.RefreshTokens("invalid-token-string")

	assert.EqualError(t, err, "invalid refresh token")
}

func TestAuthService_RefreshTokens_NotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	refreshToken, _ := auth.GenerateRefreshToken(userID)
	tokenHash := hashToken(refreshToken)

	mockTokenRepo.On("FindByTokenHash", tokenHash).Return(models.RefreshToken{}, gorm.ErrRecordNotFound)

	_, err := svc.RefreshTokens(refreshToken)

	assert.EqualError(t, err, "refresh token not found")
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_Logout_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "some-refresh-token"
	expectedHash := hashToken(rawToken)

	mockTokenRepo.On("DeleteByTokenHash", expectedHash).Return(nil)

	err := svc.Logout(rawToken)

	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_GetMe_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	user := models.User{ID: userID, Username: "testuser", Email: "test@example.com"}

	mockUserRepo.On("FindByID", userID).Return(user, nil)

	resp, err := svc.GetMe(userID)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", resp.Username)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetMe_NotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()

	mockUserRepo.On("FindByID", userID).Return(models.User{}, gorm.ErrRecordNotFound)

	_, err := svc.GetMe(userID)

	assert.Error(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_InvalidateAllSessions_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()

	mockTokenRepo.On("DeleteAllForUser", userID).Return(nil)

	err := svc.InvalidateAllSessions(userID)

	assert.NoError(t, err)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_RequestPasswordReset_UserExists(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	user := models.User{ID: userID, Email: "test@example.com"}

	mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)
	mockResetRepo.On("Create", mock.AnythingOfType("*models.PasswordResetToken")).Return(nil)
	mockResetRepo.On("DeleteExpiredForUser", userID).Return(nil)

	token, err := svc.RequestPasswordReset("test@example.com")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockUserRepo.AssertExpectations(t)
	mockResetRepo.AssertExpectations(t)
}

func TestAuthService_RequestPasswordReset_UserNotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	mockUserRepo.On("FindByEmail", "nope@example.com").Return(models.User{}, gorm.ErrRecordNotFound)

	token, err := svc.RequestPasswordReset("nope@example.com")

	// Returns empty string with nil error (silent failure for security)
	assert.NoError(t, err)
	assert.Empty(t, token)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "validresettoken"
	hash := hashToken(rawToken)
	userID := uuid.New()
	record := models.PasswordResetToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mockResetRepo.On("FindByTokenHash", hash).Return(record, nil)
	mockUserRepo.On("Update", userID, mock.AnythingOfType("map[string]interface {}")).Return(models.User{}, nil)
	mockResetRepo.On("MarkUsed", hash).Return(nil)
	mockTokenRepo.On("DeleteAllForUser", userID).Return(nil)

	err := svc.ResetPassword(rawToken, "newpassword123")

	assert.NoError(t, err)
	mockResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "badtoken"
	hash := hashToken(rawToken)

	mockResetRepo.On("FindByTokenHash", hash).Return(models.PasswordResetToken{}, gorm.ErrRecordNotFound)

	err := svc.ResetPassword(rawToken, "newpass")

	assert.EqualError(t, err, "invalid or expired reset token")
	mockResetRepo.AssertExpectations(t)
}

func TestAuthService_Login_DBError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	dbErr := errors.New("connection refused")
	mockUserRepo.On("FindByUsername", "testuser").Return(models.User{}, dbErr)

	_, _, err := svc.Login(models.LoginRequest{Username: "testuser", Password: "pass"})

	assert.ErrorIs(t, err, dbErr)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_RefreshTokens_Expired(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	refreshToken, _ := auth.GenerateRefreshToken(userID)
	tokenHash := hashToken(refreshToken)

	record := models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
	}

	mockTokenRepo.On("FindByTokenHash", tokenHash).Return(record, nil)
	// best-effort cleanup call — return value ignored by service
	mockTokenRepo.On("DeleteByTokenHash", tokenHash).Return(nil)

	_, err := svc.RefreshTokens(refreshToken)

	assert.EqualError(t, err, "refresh token expired")
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_RefreshTokens_DeleteByTokenHashError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	refreshToken, _ := auth.GenerateRefreshToken(userID)
	tokenHash := hashToken(refreshToken)

	record := models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	deleteErr := errors.New("db delete error")

	mockTokenRepo.On("FindByTokenHash", tokenHash).Return(record, nil)
	mockTokenRepo.On("DeleteByTokenHash", tokenHash).Return(deleteErr)

	_, err := svc.RefreshTokens(refreshToken)

	assert.ErrorIs(t, err, deleteErr)
	mockTokenRepo.AssertExpectations(t)
}

func TestAuthService_RequestPasswordReset_CreateError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	userID := uuid.New()
	user := models.User{ID: userID, Email: "test@example.com"}
	createErr := errors.New("db insert error")

	mockUserRepo.On("FindByEmail", "test@example.com").Return(user, nil)
	mockResetRepo.On("Create", mock.AnythingOfType("*models.PasswordResetToken")).Return(createErr)

	_, err := svc.RequestPasswordReset("test@example.com")

	assert.ErrorIs(t, err, createErr)
	mockUserRepo.AssertExpectations(t)
	mockResetRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_AlreadyUsed(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "usedtoken"
	hash := hashToken(rawToken)
	usedAt := time.Now().Add(-30 * time.Minute)
	record := models.PasswordResetToken{
		TokenHash: hash,
		UsedAt:    &usedAt,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mockResetRepo.On("FindByTokenHash", hash).Return(record, nil)

	err := svc.ResetPassword(rawToken, "newpass")

	assert.EqualError(t, err, "reset token already used")
	mockResetRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_Expired(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "expiredtoken"
	hash := hashToken(rawToken)
	record := models.PasswordResetToken{
		TokenHash: hash,
		UsedAt:    nil,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
	}

	mockResetRepo.On("FindByTokenHash", hash).Return(record, nil)

	err := svc.ResetPassword(rawToken, "newpass")

	assert.EqualError(t, err, "reset token expired")
	mockResetRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_UpdateError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "validtoken"
	hash := hashToken(rawToken)
	userID := uuid.New()
	record := models.PasswordResetToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	updateErr := errors.New("update failed")

	mockResetRepo.On("FindByTokenHash", hash).Return(record, nil)
	mockUserRepo.On("Update", userID, mock.AnythingOfType("map[string]interface {}")).Return(models.User{}, updateErr)

	err := svc.ResetPassword(rawToken, "newpass")

	assert.ErrorIs(t, err, updateErr)
	mockResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_MarkUsedError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "validtoken2"
	hash := hashToken(rawToken)
	userID := uuid.New()
	record := models.PasswordResetToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	markErr := errors.New("mark used failed")

	mockResetRepo.On("FindByTokenHash", hash).Return(record, nil)
	mockUserRepo.On("Update", userID, mock.AnythingOfType("map[string]interface {}")).Return(models.User{}, nil)
	mockResetRepo.On("MarkUsed", hash).Return(markErr)

	err := svc.ResetPassword(rawToken, "newpass")

	assert.ErrorIs(t, err, markErr)
	mockResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_DeleteAllForUserError(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTokenRepo := &mocks.MockRefreshTokenRepository{}
	mockResetRepo := &mocks.MockPasswordResetTokenRepository{}
	svc := NewAuthService(mockUserRepo, mockTokenRepo, mockResetRepo)

	rawToken := "validtoken3"
	hash := hashToken(rawToken)
	userID := uuid.New()
	record := models.PasswordResetToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	deleteErr := errors.New("delete all failed")

	mockResetRepo.On("FindByTokenHash", hash).Return(record, nil)
	mockUserRepo.On("Update", userID, mock.AnythingOfType("map[string]interface {}")).Return(models.User{}, nil)
	mockResetRepo.On("MarkUsed", hash).Return(nil)
	mockTokenRepo.On("DeleteAllForUser", userID).Return(deleteErr)

	err := svc.ResetPassword(rawToken, "newpass")

	assert.ErrorIs(t, err, deleteErr)
	mockResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}
