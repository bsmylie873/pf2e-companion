package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type PinRepository interface {
	Create(pin *models.SessionPin) error
	FindByGameID(gameID uuid.UUID) ([]models.SessionPin, error)
	FindByID(id uuid.UUID) (models.SessionPin, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.SessionPin, error)
	Delete(id uuid.UUID) error
}

type pinRepository struct {
	db *gorm.DB
}

func NewPinRepository(db *gorm.DB) PinRepository {
	return &pinRepository{db: db}
}

func (r *pinRepository) Create(pin *models.SessionPin) error {
	return r.db.Create(pin).Error
}

func (r *pinRepository) FindByGameID(gameID uuid.UUID) ([]models.SessionPin, error) {
	var pins []models.SessionPin
	err := r.db.Joins("JOIN sessions ON sessions.id = session_pins.session_id").Where("sessions.game_id = ?", gameID).Find(&pins).Error
	return pins, err
}

func (r *pinRepository) FindByID(id uuid.UUID) (models.SessionPin, error) {
	var pin models.SessionPin
	err := r.db.First(&pin, "id = ?", id).Error
	return pin, err
}

func (r *pinRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.SessionPin, error) {
	if err := r.db.Model(&models.SessionPin{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.SessionPin{}, err
	}
	return r.FindByID(id)
}

func (r *pinRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.SessionPin{}, "id = ?", id).Error
}
