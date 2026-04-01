package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

// MapRepository defines the persistence interface for GameMap records.
type MapRepository interface {
	Create(m *models.GameMap) error
	FindByID(id uuid.UUID) (models.GameMap, error)
	FindByGameID(gameID uuid.UUID) ([]models.GameMap, error)
	FindActiveByGameID(gameID uuid.UUID) ([]models.GameMap, error)
	FindArchivedByGameID(gameID uuid.UUID) ([]models.GameMap, error)
	FindByGameIDAndName(gameID uuid.UUID, name string) (models.GameMap, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.GameMap, error)
	Archive(id uuid.UUID) error
	HardDelete(id uuid.UUID) error
	FindExpiredArchived(cutoff time.Time) ([]models.GameMap, error)
}

type mapRepository struct {
	db *gorm.DB
}

// NewMapRepository constructs a MapRepository backed by the given DB connection.
func NewMapRepository(db *gorm.DB) MapRepository {
	return &mapRepository{db: db}
}

func (r *mapRepository) Create(m *models.GameMap) error {
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	return r.db.First(m, "id = ?", m.ID).Error
}

func (r *mapRepository) FindByID(id uuid.UUID) (models.GameMap, error) {
	var m models.GameMap
	err := r.db.First(&m, "id = ?", id).Error
	return m, err
}

func (r *mapRepository) FindByGameID(gameID uuid.UUID) ([]models.GameMap, error) {
	var maps []models.GameMap
	err := r.db.Where("game_id = ?", gameID).Order("sort_order ASC").Find(&maps).Error
	return maps, err
}

func (r *mapRepository) FindActiveByGameID(gameID uuid.UUID) ([]models.GameMap, error) {
	var maps []models.GameMap
	err := r.db.Where("game_id = ? AND archived_at IS NULL", gameID).Order("sort_order ASC").Find(&maps).Error
	return maps, err
}

func (r *mapRepository) FindArchivedByGameID(gameID uuid.UUID) ([]models.GameMap, error) {
	var maps []models.GameMap
	err := r.db.Where("game_id = ? AND archived_at IS NOT NULL AND archived_at > ?", gameID, time.Now().Add(-24*time.Hour)).Order("archived_at DESC").Find(&maps).Error
	return maps, err
}

func (r *mapRepository) FindByGameIDAndName(gameID uuid.UUID, name string) (models.GameMap, error) {
	var m models.GameMap
	err := r.db.First(&m, "game_id = ? AND name = ? AND archived_at IS NULL", gameID, name).Error
	return m, err
}

func (r *mapRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.GameMap, error) {
	if err := r.db.Model(&models.GameMap{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.GameMap{}, err
	}
	return r.FindByID(id)
}

func (r *mapRepository) Archive(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.GameMap{}).Where("id = ?", id).Update("archived_at", now).Error
}

func (r *mapRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.GameMap{}, "id = ?", id).Error
}

func (r *mapRepository) FindExpiredArchived(cutoff time.Time) ([]models.GameMap, error) {
	var maps []models.GameMap
	err := r.db.Where("archived_at IS NOT NULL AND archived_at <= ?", cutoff).Find(&maps).Error
	return maps, err
}
