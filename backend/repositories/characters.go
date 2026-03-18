package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type CharacterRepository interface {
	Create(character *models.Character) error
	FindByGameID(gameID uuid.UUID) ([]models.Character, error)
	FindByID(id uuid.UUID) (models.Character, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Character, error)
	Delete(id uuid.UUID) error
}

type characterRepository struct {
	db *gorm.DB
}

func NewCharacterRepository(db *gorm.DB) CharacterRepository {
	return &characterRepository{db: db}
}

func (r *characterRepository) Create(character *models.Character) error {
	return r.db.Create(character).Error
}

func (r *characterRepository) FindByGameID(gameID uuid.UUID) ([]models.Character, error) {
	var characters []models.Character
	err := r.db.Where("game_id = ?", gameID).Find(&characters).Error
	return characters, err
}

func (r *characterRepository) FindByID(id uuid.UUID) (models.Character, error) {
	var character models.Character
	err := r.db.First(&character, "id = ?", id).Error
	return character, err
}

func (r *characterRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Character, error) {
	if err := r.db.Model(&models.Character{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Character{}, err
	}
	return r.FindByID(id)
}

func (r *characterRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Character{}, "id = ?", id).Error
}
