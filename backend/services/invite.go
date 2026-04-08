package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

// InviteService handles generation, validation, and redemption of magic-link invite tokens.
type InviteService interface {
	GenerateInvite(gameID, callerID uuid.UUID, expiresIn string) (models.InviteTokenResponse, error)
	GetActiveInvite(gameID, callerID uuid.UUID) (models.InviteTokenStatusResponse, error)
	RevokeInvite(gameID, callerID uuid.UUID) error
	ValidateInvite(rawToken string) (models.InviteValidationResponse, error)
	RedeemInvite(rawToken string, userID uuid.UUID) (models.InviteRedeemResponse, error)
}

type inviteService struct {
	inviteRepo     repositories.InviteTokenRepository
	membershipRepo repositories.MembershipRepository
	gameRepo       repositories.GameRepository
}

// NewInviteService creates a new InviteService.
func NewInviteService(
	inviteRepo repositories.InviteTokenRepository,
	membershipRepo repositories.MembershipRepository,
	gameRepo repositories.GameRepository,
) InviteService {
	return &inviteService{
		inviteRepo:     inviteRepo,
		membershipRepo: membershipRepo,
		gameRepo:       gameRepo,
	}
}

func hashInviteToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *inviteService) requireGM(callerID, gameID uuid.UUID) error {
	m, err := s.membershipRepo.FindByUserAndGameID(callerID, gameID)
	if err != nil {
		return ErrForbidden
	}
	if !m.IsGM {
		return ErrForbidden
	}
	return nil
}

func parseExpiresIn(expiresIn string) (*time.Time, error) {
	switch expiresIn {
	case "never", "":
		return nil, nil
	case "24h":
		t := time.Now().Add(24 * time.Hour)
		return &t, nil
	case "7d":
		t := time.Now().Add(7 * 24 * time.Hour)
		return &t, nil
	default:
		return nil, errors.New("invalid expires_in value; use \"24h\", \"7d\", or \"never\"")
	}
}

func (s *inviteService) GenerateInvite(gameID, callerID uuid.UUID, expiresIn string) (models.InviteTokenResponse, error) {
	if err := s.requireGM(callerID, gameID); err != nil {
		return models.InviteTokenResponse{}, err
	}

	expiresAt, err := parseExpiresIn(expiresIn)
	if err != nil {
		return models.InviteTokenResponse{}, ErrValidation
	}

	// Revoke any existing active invite for this game
	_ = s.inviteRepo.RevokeAllForGame(gameID)

	// Generate 32-byte random token, base64url encoded
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return models.InviteTokenResponse{}, err
	}
	rawToken := base64.URLEncoding.EncodeToString(buf)

	record := models.InviteToken{
		GameID:    gameID,
		CreatedBy: callerID,
		TokenHash: hashInviteToken(rawToken),
		ExpiresAt: expiresAt,
	}
	if err := s.inviteRepo.Create(&record); err != nil {
		return models.InviteTokenResponse{}, err
	}

	return models.InviteTokenResponse{
		Token:     rawToken,
		ExpiresAt: record.ExpiresAt,
		CreatedAt: record.CreatedAt,
	}, nil
}

func (s *inviteService) GetActiveInvite(gameID, callerID uuid.UUID) (models.InviteTokenStatusResponse, error) {
	if err := s.requireGM(callerID, gameID); err != nil {
		return models.InviteTokenStatusResponse{}, err
	}

	token, err := s.inviteRepo.FindActiveByGameID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.InviteTokenStatusResponse{HasActiveInvite: false}, nil
		}
		return models.InviteTokenStatusResponse{}, err
	}

	return models.InviteTokenStatusResponse{
		HasActiveInvite: true,
		ExpiresAt:       token.ExpiresAt,
		CreatedAt:       &token.CreatedAt,
	}, nil
}

func (s *inviteService) RevokeInvite(gameID, callerID uuid.UUID) error {
	if err := s.requireGM(callerID, gameID); err != nil {
		return err
	}
	return s.inviteRepo.RevokeAllForGame(gameID)
}

func (s *inviteService) ValidateInvite(rawToken string) (models.InviteValidationResponse, error) {
	hash := hashInviteToken(rawToken)
	token, err := s.inviteRepo.FindByTokenHash(hash)
	if err != nil {
		return models.InviteValidationResponse{}, errors.New("invalid invite token")
	}

	if token.RevokedAt != nil {
		return models.InviteValidationResponse{}, errors.New("invite token has been revoked")
	}
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return models.InviteValidationResponse{}, errors.New("invite token has expired")
	}

	game, err := s.gameRepo.FindByID(token.GameID)
	if err != nil {
		return models.InviteValidationResponse{}, errors.New("game not found")
	}

	return models.InviteValidationResponse{
		GameID:    game.ID,
		GameTitle: game.Title,
	}, nil
}

func (s *inviteService) RedeemInvite(rawToken string, userID uuid.UUID) (models.InviteRedeemResponse, error) {
	hash := hashInviteToken(rawToken)
	token, err := s.inviteRepo.FindByTokenHash(hash)
	if err != nil {
		return models.InviteRedeemResponse{}, errors.New("invalid invite token")
	}

	if token.RevokedAt != nil {
		return models.InviteRedeemResponse{}, errors.New("invite token has been revoked")
	}
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return models.InviteRedeemResponse{}, errors.New("invite token has expired")
	}

	// Check if user is already a member
	existing, err := s.membershipRepo.FindByUserAndGameID(userID, token.GameID)
	if err == nil {
		// Already a member — return success with already_member flag
		return models.InviteRedeemResponse{
			GameID:        token.GameID,
			MembershipID:  existing.ID,
			AlreadyMember: true,
		}, nil
	}

	// Create membership
	membership := models.GameMembership{
		GameID: token.GameID,
		UserID: userID,
		IsGM:   false,
	}
	if err := s.membershipRepo.Create(&membership); err != nil {
		return models.InviteRedeemResponse{}, err
	}

	return models.InviteRedeemResponse{
		GameID:        token.GameID,
		MembershipID:  membership.ID,
		AlreadyMember: false,
	}, nil
}
