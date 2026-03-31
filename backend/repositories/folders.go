package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type FolderRepository interface {
	Create(folder *models.Folder) error
	FindByID(id uuid.UUID) (models.Folder, error)
	FindSessionFolders(gameID uuid.UUID) ([]models.Folder, error)
	FindNoteFolders(gameID, userID uuid.UUID) ([]models.Folder, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Folder, error)
	Delete(id uuid.UUID) error
	BatchUpdatePositions(ids []uuid.UUID, positions []int) error
	MaxPosition(gameID uuid.UUID, folderType string, userID *uuid.UUID) (int, error)
}

type folderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) FolderRepository {
	return &folderRepository{db: db}
}

func (r *folderRepository) Create(folder *models.Folder) error {
	return r.db.Create(folder).Error
}

func (r *folderRepository) FindByID(id uuid.UUID) (models.Folder, error) {
	var folder models.Folder
	err := r.db.First(&folder, "id = ?", id).Error
	return folder, err
}

func (r *folderRepository) FindSessionFolders(gameID uuid.UUID) ([]models.Folder, error) {
	var folders []models.Folder
	err := r.db.Where("game_id = ? AND folder_type = 'session'", gameID).
		Order("position ASC").
		Find(&folders).Error
	return folders, err
}

func (r *folderRepository) FindNoteFolders(gameID, userID uuid.UUID) ([]models.Folder, error) {
	var folders []models.Folder
	err := r.db.Where("game_id = ? AND folder_type = 'note' AND user_id = ?", gameID, userID).
		Order("position ASC").
		Find(&folders).Error
	return folders, err
}

func (r *folderRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Folder, error) {
	if err := r.db.Model(&models.Folder{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Folder{}, err
	}
	return r.FindByID(id)
}

func (r *folderRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Folder{}, "id = ?", id).Error
}

func (r *folderRepository) BatchUpdatePositions(ids []uuid.UUID, positions []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&models.Folder{}).Where("id = ?", id).Update("position", positions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *folderRepository) MaxPosition(gameID uuid.UUID, folderType string, userID *uuid.UUID) (int, error) {
	var maxPos int
	query := r.db.Model(&models.Folder{}).Where("game_id = ? AND folder_type = ?", gameID, folderType)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	err := query.Select("COALESCE(MAX(position), -1)").Row().Scan(&maxPos)
	if err != nil {
		return -1, err
	}
	return maxPos, nil
}
