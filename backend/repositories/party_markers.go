package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type PartyMarkerRepository interface {
	FindByGameID(gameID uuid.UUID) (*models.PartyMarker, error)
	Upsert(marker *models.PartyMarker) error
	Delete(gameID uuid.UUID) error
}

type partyMarkerRepository struct {
	db *gorm.DB
}

func NewPartyMarkerRepository(db *gorm.DB) PartyMarkerRepository {
	return &partyMarkerRepository{db: db}
}

func (r *partyMarkerRepository) FindByGameID(gameID uuid.UUID) (*models.PartyMarker, error) {
	var marker models.PartyMarker
	result := r.db.First(&marker, "game_id = ?", gameID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &marker, nil
}

func (r *partyMarkerRepository) Upsert(marker *models.PartyMarker) error {
	err := r.db.Exec(
		`INSERT INTO party_markers (game_id, map_id, x, y) VALUES (?, ?, ?, ?) ON CONFLICT (game_id) DO UPDATE SET map_id = EXCLUDED.map_id, x = EXCLUDED.x, y = EXCLUDED.y, updated_at = now()`,
		marker.GameID, marker.MapID, marker.X, marker.Y,
	).Error
	if err != nil {
		return err
	}
	return r.db.First(marker, "game_id = ?", marker.GameID).Error
}

func (r *partyMarkerRepository) Delete(gameID uuid.UUID) error {
	return r.db.Delete(&models.PartyMarker{}, "game_id = ?", gameID).Error
}
