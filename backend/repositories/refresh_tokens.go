package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

// RefreshTokenRepository provides CRUD operations for refresh tokens.
type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByTokenHash(hash string) (models.RefreshToken, error)
	DeleteByTokenHash(hash string) error
	DeleteExpiredForUser(userID uuid.UUID) error
	DeleteAllForUser(userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository backed by GORM.
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(hash string) (models.RefreshToken, error) {
	var t models.RefreshToken
	err := r.db.First(&t, "token_hash = ?", hash).Error
	return t, err
}

func (r *refreshTokenRepository) DeleteByTokenHash(hash string) error {
	return r.db.Delete(&models.RefreshToken{}, "token_hash = ?", hash).Error
}

func (r *refreshTokenRepository) DeleteExpiredForUser(userID uuid.UUID) error {
	return r.db.Delete(&models.RefreshToken{}, "user_id = ? AND expires_at < ?", userID, time.Now()).Error
}

func (r *refreshTokenRepository) DeleteAllForUser(userID uuid.UUID) error {
	return r.db.Delete(&models.RefreshToken{}, "user_id = ?", userID).Error
}
