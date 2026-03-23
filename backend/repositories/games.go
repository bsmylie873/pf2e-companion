package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type GameRepository interface {
	Create(game *models.Game) error
	FindAll() ([]models.Game, error)
	FindByID(id uuid.UUID) (models.Game, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Game, error)
	Delete(id uuid.UUID) error
	FindByIDs(ids []uuid.UUID) ([]models.Game, error)
}

type gameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) GameRepository {
	return &gameRepository{db: db}
}

func (r *gameRepository) Create(game *models.Game) error {
	return r.db.Create(game).Error
}

func (r *gameRepository) FindAll() ([]models.Game, error) {
	var games []models.Game
	err := r.db.Find(&games).Error
	return games, err
}

func (r *gameRepository) FindByID(id uuid.UUID) (models.Game, error) {
	var game models.Game
	err := r.db.First(&game, "id = ?", id).Error
	return game, err
}

func (r *gameRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Game, error) {
	if err := r.db.Model(&models.Game{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Game{}, err
	}
	return r.FindByID(id)
}

func (r *gameRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Game{}, "id = ?", id).Error
}

func (r *gameRepository) FindByIDs(ids []uuid.UUID) ([]models.Game, error) {
	if len(ids) == 0 {
		return []models.Game{}, nil
	}
	var games []models.Game
	err := r.db.Where("id IN ?", ids).Find(&games).Error
	return games, err
}
