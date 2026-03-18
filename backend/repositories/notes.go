package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type NoteRepository interface {
	Create(note *models.Note) error
	FindByGameID(gameID uuid.UUID) ([]models.Note, error)
	FindByUserID(userID uuid.UUID) ([]models.Note, error)
	FindByID(id uuid.UUID) (models.Note, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Note, error)
	Delete(id uuid.UUID) error
}

type noteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(note *models.Note) error {
	return r.db.Create(note).Error
}

func (r *noteRepository) FindByGameID(gameID uuid.UUID) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Where("game_id = ?", gameID).Find(&notes).Error
	return notes, err
}

func (r *noteRepository) FindByUserID(userID uuid.UUID) ([]models.Note, error) {
	var notes []models.Note
	err := r.db.Where("user_id = ?", userID).Find(&notes).Error
	return notes, err
}

func (r *noteRepository) FindByID(id uuid.UUID) (models.Note, error) {
	var note models.Note
	err := r.db.First(&note, "id = ?", id).Error
	return note, err
}

func (r *noteRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	if err := r.db.Model(&models.Note{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Note{}, err
	}
	return r.FindByID(id)
}

func (r *noteRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Note{}, "id = ?", id).Error
}
