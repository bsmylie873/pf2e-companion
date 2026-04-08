package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

// InviteTokenRepository provides CRUD operations for invite tokens.
type InviteTokenRepository interface {
	Create(token *models.InviteToken) error
	FindActiveByGameID(gameID uuid.UUID) (models.InviteToken, error)
	FindByTokenHash(hash string) (models.InviteToken, error)
	RevokeAllForGame(gameID uuid.UUID) error
	RevokeByID(id uuid.UUID) error
}

type inviteTokenRepository struct {
	db *gorm.DB
}

// NewInviteTokenRepository creates a new InviteTokenRepository backed by GORM.
func NewInviteTokenRepository(db *gorm.DB) InviteTokenRepository {
	return &inviteTokenRepository{db: db}
}

func (r *inviteTokenRepository) Create(token *models.InviteToken) error {
	return r.db.Create(token).Error
}

func (r *inviteTokenRepository) FindActiveByGameID(gameID uuid.UUID) (models.InviteToken, error) {
	var t models.InviteToken
	err := r.db.First(&t, "game_id = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > NOW())", gameID).Error
	return t, err
}

func (r *inviteTokenRepository) FindByTokenHash(hash string) (models.InviteToken, error) {
	var t models.InviteToken
	err := r.db.First(&t, "token_hash = ?", hash).Error
	return t, err
}

func (r *inviteTokenRepository) RevokeAllForGame(gameID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.InviteToken{}).Where("game_id = ? AND revoked_at IS NULL", gameID).Update("revoked_at", now).Error
}

func (r *inviteTokenRepository) RevokeByID(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.InviteToken{}).Where("id = ?", id).Update("revoked_at", now).Error
}
