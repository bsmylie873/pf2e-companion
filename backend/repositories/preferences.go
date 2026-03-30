package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type PreferenceRepository interface {
	FindByUserID(userID uuid.UUID) (models.UserPreference, error)
	Upsert(pref *models.UserPreference) error
	ClearDefaultGameForGame(gameID uuid.UUID) error
	ClearDefaultGameForMembership(userID, gameID uuid.UUID) error
}

type preferenceRepository struct {
	db *gorm.DB
}

func NewPreferenceRepository(db *gorm.DB) PreferenceRepository {
	return &preferenceRepository{db: db}
}

func (r *preferenceRepository) FindByUserID(userID uuid.UUID) (models.UserPreference, error) {
	var pref models.UserPreference
	err := r.db.First(&pref, "user_id = ?", userID).Error
	return pref, err
}

func (r *preferenceRepository) Upsert(pref *models.UserPreference) error {
	return r.db.Save(pref).Error
}

func (r *preferenceRepository) ClearDefaultGameForGame(gameID uuid.UUID) error {
	return r.db.Model(&models.UserPreference{}).
		Where("default_game_id = ?", gameID).
		Update("default_game_id", nil).Error
}

func (r *preferenceRepository) ClearDefaultGameForMembership(userID, gameID uuid.UUID) error {
	return r.db.Model(&models.UserPreference{}).
		Where("user_id = ? AND default_game_id = ?", userID, gameID).
		Update("default_game_id", nil).Error
}
