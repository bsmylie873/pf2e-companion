package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type ItemRepository interface {
	Create(item *models.Item) error
	FindByGameID(gameID uuid.UUID) ([]models.Item, error)
	FindByCharacterID(characterID uuid.UUID) ([]models.Item, error)
	FindByID(id uuid.UUID) (models.Item, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Item, error)
	Delete(id uuid.UUID) error
}

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Create(item *models.Item) error {
	return r.db.Create(item).Error
}

func (r *itemRepository) FindByGameID(gameID uuid.UUID) ([]models.Item, error) {
	var items []models.Item
	err := r.db.Where("game_id = ?", gameID).Find(&items).Error
	return items, err
}

func (r *itemRepository) FindByCharacterID(characterID uuid.UUID) ([]models.Item, error) {
	var items []models.Item
	err := r.db.Where("character_id = ?", characterID).Find(&items).Error
	return items, err
}

func (r *itemRepository) FindByID(id uuid.UUID) (models.Item, error) {
	var item models.Item
	err := r.db.First(&item, "id = ?", id).Error
	return item, err
}

func (r *itemRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Item, error) {
	if err := r.db.Model(&models.Item{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Item{}, err
	}
	return r.FindByID(id)
}

func (r *itemRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Item{}, "id = ?", id).Error
}
