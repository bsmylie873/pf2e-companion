package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
)

type SessionRepository interface {
	Create(session *models.Session) error
	FindByGameID(gameID uuid.UUID) ([]models.Session, error)
	FindByGameIDPaginated(gameID uuid.UUID, offset, limit int) ([]models.Session, int64, error)
	FindByID(id uuid.UUID) (models.Session, error)
	Update(id uuid.UUID, updates map[string]interface{}) (models.Session, error)
	Delete(id uuid.UUID) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(session *models.Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) FindByGameID(gameID uuid.UUID) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.Where("game_id = ?", gameID).Order("created_at DESC").Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) FindByGameIDPaginated(gameID uuid.UUID, offset, limit int) ([]models.Session, int64, error) {
	var count int64
	var sessions []models.Session
	baseQuery := r.db.Model(&models.Session{}).Where("game_id = ?", gameID)
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := baseQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&sessions).Error
	return sessions, count, err
}

func (r *sessionRepository) FindByID(id uuid.UUID) (models.Session, error) {
	var session models.Session
	err := r.db.First(&session, "id = ?", id).Error
	return session, err
}

func (r *sessionRepository) Update(id uuid.UUID, updates map[string]interface{}) (models.Session, error) {
	updates["version"] = gorm.Expr("version + 1")
	if err := r.db.Model(&models.Session{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return models.Session{}, err
	}
	return r.FindByID(id)
}

func (r *sessionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Session{}, "id = ?", id).Error
}
