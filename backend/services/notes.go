package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type NoteService interface {
	CreateGameNote(gameID uuid.UUID, note *models.Note) (models.Note, error)
	CreateUserNote(userID uuid.UUID, note *models.Note) (models.Note, error)
	ListGameNotes(gameID uuid.UUID) ([]models.Note, error)
	ListUserNotes(userID uuid.UUID) ([]models.Note, error)
	GetNote(id uuid.UUID) (models.Note, error)
	UpdateNote(id uuid.UUID, updates map[string]interface{}) (models.Note, error)
	DeleteNote(id uuid.UUID) error
}

type noteService struct {
	repo repositories.NoteRepository
}

func NewNoteService(repo repositories.NoteRepository) NoteService {
	return &noteService{repo: repo}
}

func (s *noteService) CreateGameNote(gameID uuid.UUID, note *models.Note) (models.Note, error) {
	note.ID = uuid.Nil
	note.GameID = &gameID
	note.UserID = nil
	if err := s.repo.Create(note); err != nil {
		return models.Note{}, err
	}
	return *note, nil
}

func (s *noteService) CreateUserNote(userID uuid.UUID, note *models.Note) (models.Note, error) {
	note.ID = uuid.Nil
	note.UserID = &userID
	note.GameID = nil
	if err := s.repo.Create(note); err != nil {
		return models.Note{}, err
	}
	return *note, nil
}

func (s *noteService) ListGameNotes(gameID uuid.UUID) ([]models.Note, error) {
	return s.repo.FindByGameID(gameID)
}

func (s *noteService) ListUserNotes(userID uuid.UUID) ([]models.Note, error) {
	return s.repo.FindByUserID(userID)
}

func (s *noteService) GetNote(id uuid.UUID) (models.Note, error) {
	return s.repo.FindByID(id)
}

func (s *noteService) UpdateNote(id uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	if _, err := s.repo.FindByID(id); err != nil {
		return models.Note{}, err
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *noteService) DeleteNote(id uuid.UUID) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
