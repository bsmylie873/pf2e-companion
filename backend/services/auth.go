package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"pf2e-companion/backend/auth"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

// AuthService handles user registration, login, token refresh, logout, and password reset.
type AuthService interface {
	Register(req models.RegisterRequest) (models.UserResponse, models.TokenPair, error)
	Login(req models.LoginRequest) (models.UserResponse, models.TokenPair, error)
	RefreshTokens(refreshToken string) (models.TokenPair, error)
	Logout(refreshToken string) error
	GetMe(userID uuid.UUID) (models.UserResponse, error)
	InvalidateAllSessions(userID uuid.UUID) error
	RequestPasswordReset(email string) error
	ResetPassword(token, newPassword string) error
}

type authService struct {
	userRepo       repositories.UserRepository
	tokenRepo      repositories.RefreshTokenRepository
	resetTokenRepo repositories.PasswordResetTokenRepository
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repositories.UserRepository, tokenRepo repositories.RefreshTokenRepository, resetTokenRepo repositories.PasswordResetTokenRepository) AuthService {
	return &authService{userRepo: userRepo, tokenRepo: tokenRepo, resetTokenRepo: resetTokenRepo}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *authService) Register(req models.RegisterRequest) (models.UserResponse, models.TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, models.TokenPair{}, err
	}
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}
	if err := s.userRepo.Create(&user); err != nil {
		return models.UserResponse{}, models.TokenPair{}, err
	}
	pair, err := s.issueTokenPair(user.ID)
	if err != nil {
		return models.UserResponse{}, models.TokenPair{}, err
	}
	return models.FromUser(user), pair, nil
}

func (s *authService) Login(req models.LoginRequest) (models.UserResponse, models.TokenPair, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.UserResponse{}, models.TokenPair{}, errors.New("invalid credentials")
		}
		return models.UserResponse{}, models.TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return models.UserResponse{}, models.TokenPair{}, errors.New("invalid credentials")
	}

	pair, err := s.issueTokenPair(user.ID)
	if err != nil {
		return models.UserResponse{}, models.TokenPair{}, err
	}

	return models.FromUser(user), pair, nil
}

func (s *authService) issueTokenPair(userID uuid.UUID) (models.TokenPair, error) {
	accessToken, err := auth.GenerateAccessToken(userID)
	if err != nil {
		return models.TokenPair{}, err
	}

	refreshToken, err := auth.GenerateRefreshToken(userID)
	if err != nil {
		return models.TokenPair{}, err
	}

	tokenRecord := models.RefreshToken{
		UserID:    userID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.tokenRepo.Create(&tokenRecord); err != nil {
		return models.TokenPair{}, err
	}

	// Best-effort cleanup of expired tokens
	_ = s.tokenRepo.DeleteExpiredForUser(userID)

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) RefreshTokens(refreshToken string) (models.TokenPair, error) {
	claims, err := auth.ValidateToken(refreshToken)
	if err != nil {
		return models.TokenPair{}, errors.New("invalid refresh token")
	}

	hash := hashToken(refreshToken)
	record, err := s.tokenRepo.FindByTokenHash(hash)
	if err != nil {
		return models.TokenPair{}, errors.New("refresh token not found")
	}

	if time.Now().After(record.ExpiresAt) {
		_ = s.tokenRepo.DeleteByTokenHash(hash)
		return models.TokenPair{}, errors.New("refresh token expired")
	}

	if err := s.tokenRepo.DeleteByTokenHash(hash); err != nil {
		return models.TokenPair{}, err
	}

	return s.issueTokenPair(claims.UserID)
}

func (s *authService) Logout(refreshToken string) error {
	hash := hashToken(refreshToken)
	return s.tokenRepo.DeleteByTokenHash(hash)
}

func (s *authService) GetMe(userID uuid.UUID) (models.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return models.UserResponse{}, err
	}
	return models.FromUser(user), nil
}

func (s *authService) InvalidateAllSessions(userID uuid.UUID) error {
	return s.tokenRepo.DeleteAllForUser(userID)
}

func (s *authService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		// Always return nil — never reveal whether the email exists
		return nil
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	rawToken := hex.EncodeToString(buf)

	record := models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := s.resetTokenRepo.Create(&record); err != nil {
		return err
	}

	// Best-effort cleanup of expired tokens
	_ = s.resetTokenRepo.DeleteExpiredForUser(user.ID)

	// No email service yet — log the raw token to stdout
	fmt.Printf("[PASSWORD RESET] reset token for %s: %s\n", email, rawToken)
	return nil
}

func (s *authService) ResetPassword(token, newPassword string) error {
	hash := hashToken(token)
	record, err := s.resetTokenRepo.FindByTokenHash(hash)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if record.UsedAt != nil {
		return errors.New("reset token already used")
	}

	if time.Now().After(record.ExpiresAt) {
		return errors.New("reset token expired")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err := s.userRepo.Update(record.UserID, map[string]interface{}{"password_hash": string(newHash)}); err != nil {
		return err
	}

	// Mark the reset token as used
	if err := s.resetTokenRepo.MarkUsed(hash); err != nil {
		return err
	}

	// Invalidate all refresh tokens to force re-login on all devices
	return s.tokenRepo.DeleteAllForUser(record.UserID)
}
