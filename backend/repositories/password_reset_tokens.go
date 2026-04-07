package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

// PasswordResetTokenRepository provides CRUD operations for password-reset tokens.
type PasswordResetTokenRepository interface {
	Create(token *models.PasswordResetToken) error
	FindByTokenHash(hash string) (models.PasswordResetToken, error)
	MarkUsed(hash string) error
	DeleteExpiredForUser(userID uuid.UUID) error
	DeleteAllForUser(userID uuid.UUID) error
}

type passwordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository creates a new PasswordResetTokenRepository backed by GORM.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(token *models.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *passwordResetTokenRepository) FindByTokenHash(hash string) (models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	err := r.db.First(&t, "token_hash = ?", hash).Error
	return t, err
}

func (r *passwordResetTokenRepository) MarkUsed(hash string) error {
	now := time.Now()
	return r.db.Model(&models.PasswordResetToken{}).Where("token_hash = ?", hash).Update("used_at", now).Error
}

func (r *passwordResetTokenRepository) DeleteExpiredForUser(userID uuid.UUID) error {
	return r.db.Delete(&models.PasswordResetToken{}, "user_id = ? AND expires_at < ?", userID, time.Now()).Error
}

func (r *passwordResetTokenRepository) DeleteAllForUser(userID uuid.UUID) error {
	return r.db.Delete(&models.PasswordResetToken{}, "user_id = ?", userID).Error
}
