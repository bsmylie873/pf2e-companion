package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type NoteFilters struct {
	Sort      string
	SessionID *uuid.UUID
	FolderID  *uuid.UUID
	Unlinked  bool
}

type NoteRepository interface {
	Create(note *models.Note) error
	FindByGameID(gameID, userID uuid.UUID, isGM bool, filters NoteFilters) ([]models.Note, error)
	FindByGameIDPaginated(gameID, userID uuid.UUID, isGM bool, filters NoteFilters, offset, limit int) ([]models.Note, int64, error)
	FindByID(id uuid.UUID) (models.Note, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Note, error)
	Delete(id uuid.UUID) error
	ClearNoteFromPins(noteID uuid.UUID) error
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

func (r *noteRepository) FindByGameID(gameID, userID uuid.UUID, isGM bool, filters NoteFilters) ([]models.Note, error) {
	var notes []models.Note
	query := r.db.Where("game_id = ?", gameID)

	if !isGM {
		query = query.Where("visibility IN ('visible', 'editable') OR user_id = ?", userID)
	}

	if filters.SessionID != nil {
		query = query.Where("session_id = ?", *filters.SessionID)
	} else if filters.Unlinked {
		query = query.Where("session_id IS NULL")
	}

	if filters.FolderID != nil {
		query = query.Where("folder_id = ?", *filters.FolderID)
	}

	if filters.Sort == "title" {
		query = query.Order("title ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	err := query.Find(&notes).Error
	return notes, err
}

func (r *noteRepository) FindByGameIDPaginated(gameID, userID uuid.UUID, isGM bool, filters NoteFilters, offset, limit int) ([]models.Note, int64, error) {
	baseQuery := r.db.Model(&models.Note{}).Where("game_id = ?", gameID)

	if !isGM {
		baseQuery = baseQuery.Where("visibility IN ('visible', 'editable') OR user_id = ?", userID)
	}

	if filters.SessionID != nil {
		baseQuery = baseQuery.Where("session_id = ?", *filters.SessionID)
	} else if filters.Unlinked {
		baseQuery = baseQuery.Where("session_id IS NULL")
	}

	if filters.FolderID != nil {
		baseQuery = baseQuery.Where("folder_id = ?", *filters.FolderID)
	}

	var count int64
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	var notes []models.Note
	if filters.Sort == "title" {
		baseQuery = baseQuery.Order("title ASC")
	} else {
		baseQuery = baseQuery.Order("created_at DESC")
	}

	err := baseQuery.Offset(offset).Limit(limit).Find(&notes).Error
	return notes, count, err
}

func (r *noteRepository) FindByID(id uuid.UUID) (models.Note, error) {
	var note models.Note
	err := r.db.First(&note, "id = ?", id).Error
	return note, err
}

func (r *noteRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	updates["version"] = gorm.Expr("version + 1")
	if err := r.db.Model(&models.Note{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Note{}, err
	}
	return r.FindByID(id)
}

func (r *noteRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Note{}, "id = ?", id).Error
}

func (r *noteRepository) ClearNoteFromPins(noteID uuid.UUID) error {
	return r.db.Model(&models.SessionPin{}).Where("note_id = ?", noteID).Update("note_id", nil).Error
}
