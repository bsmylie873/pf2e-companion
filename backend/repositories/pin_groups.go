package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type PinGroupRepository interface {
	Create(group *models.PinGroup) error
	FindByID(id uuid.UUID) (models.PinGroup, error)
	FindByGameID(gameID uuid.UUID) ([]models.PinGroup, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.PinGroup, error)
	Delete(id uuid.UUID) error
	CountMembers(groupID uuid.UUID) (int64, error)
}

type pinGroupRepository struct {
	db *gorm.DB
}

func NewPinGroupRepository(db *gorm.DB) PinGroupRepository {
	return &pinGroupRepository{db: db}
}

func (r *pinGroupRepository) Create(group *models.PinGroup) error {
	if err := r.db.Create(group).Error; err != nil {
		return err
	}
	return r.db.First(group, "id = ?", group.ID).Error
}

func (r *pinGroupRepository) FindByID(id uuid.UUID) (models.PinGroup, error) {
	var group models.PinGroup
	err := r.db.First(&group, "id = ?", id).Error
	return group, err
}

func (r *pinGroupRepository) FindByGameID(gameID uuid.UUID) ([]models.PinGroup, error) {
	var groups []models.PinGroup
	err := r.db.Where("game_id = ?", gameID).Find(&groups).Error
	return groups, err
}

func (r *pinGroupRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.PinGroup, error) {
	if err := r.db.Model(&models.PinGroup{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.PinGroup{}, err
	}
	return r.FindByID(id)
}

func (r *pinGroupRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.PinGroup{}, "id = ?", id).Error
}

func (r *pinGroupRepository) CountMembers(groupID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.SessionPin{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}
