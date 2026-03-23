package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type MembershipRepository interface {
	Create(membership *models.GameMembership) error
	FindByGameID(gameID uuid.UUID) ([]models.GameMembership, error)
	FindByID(id uuid.UUID) (models.GameMembership, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.GameMembership, error)
	Delete(id uuid.UUID) error
	FindByUserID(userID uuid.UUID) ([]models.GameMembership, error)
	FindByUserAndGameID(userID, gameID uuid.UUID) (models.GameMembership, error)
}

type membershipRepository struct {
	db *gorm.DB
}

func NewMembershipRepository(db *gorm.DB) MembershipRepository {
	return &membershipRepository{db: db}
}

func (r *membershipRepository) Create(membership *models.GameMembership) error {
	return r.db.Create(membership).Error
}

func (r *membershipRepository) FindByGameID(gameID uuid.UUID) ([]models.GameMembership, error) {
	var memberships []models.GameMembership
	err := r.db.Where("game_id = ?", gameID).Find(&memberships).Error
	return memberships, err
}

func (r *membershipRepository) FindByID(id uuid.UUID) (models.GameMembership, error) {
	var membership models.GameMembership
	err := r.db.First(&membership, "id = ?", id).Error
	return membership, err
}

func (r *membershipRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.GameMembership, error) {
	if err := r.db.Model(&models.GameMembership{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.GameMembership{}, err
	}
	return r.FindByID(id)
}

func (r *membershipRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.GameMembership{}, "id = ?", id).Error
}

func (r *membershipRepository) FindByUserID(userID uuid.UUID) ([]models.GameMembership, error) {
	var memberships []models.GameMembership
	err := r.db.Where("user_id = ?", userID).Find(&memberships).Error
	return memberships, err
}

func (r *membershipRepository) FindByUserAndGameID(userID, gameID uuid.UUID) (models.GameMembership, error) {
	var m models.GameMembership
	err := r.db.First(&m, "user_id = ? AND game_id = ?", userID, gameID).Error
	return m, err
}
