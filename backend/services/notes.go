package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type NoteService interface {
	CreateNote(gameID, userID uuid.UUID, note *models.Note) (models.Note, error)
	ListGameNotes(gameID, userID uuid.UUID, filters repositories.NoteFilters) ([]models.Note, error)
	GetNote(id, userID uuid.UUID) (models.Note, error)
	UpdateNote(id, userID uuid.UUID, updates map[string]interface{}) (models.Note, error)
	DeleteNote(id, userID uuid.UUID) error
}

type noteService struct {
	repo           repositories.NoteRepository
	membershipRepo repositories.MembershipRepository
}

func NewNoteService(repo repositories.NoteRepository, membershipRepo repositories.MembershipRepository) NoteService {
	return &noteService{repo: repo, membershipRepo: membershipRepo}
}

func (s *noteService) CreateNote(gameID, userID uuid.UUID, note *models.Note) (models.Note, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.Note{}, ErrForbidden
	}
	note.ID = uuid.Nil
	note.GameID = gameID
	note.UserID = userID
	if note.Visibility == "" {
		note.Visibility = "private"
	}
	if err := s.repo.Create(note); err != nil {
		return models.Note{}, err
	}
	return *note, nil
}

func (s *noteService) ListGameNotes(gameID, userID uuid.UUID, filters repositories.NoteFilters) ([]models.Note, error) {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, gameID)
	if err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID, userID, membership.IsGM, filters)
}

func (s *noteService) GetNote(id, userID uuid.UUID) (models.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return models.Note{}, err
	}
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, note.GameID)
	if err != nil {
		return models.Note{}, ErrForbidden
	}
	if membership.IsGM {
		return note, nil
	}
	if note.Visibility == "shared" {
		return note, nil
	}
	if note.UserID == userID {
		return note, nil
	}
	return models.Note{}, ErrForbidden
}

func (s *noteService) UpdateNote(id, userID uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return models.Note{}, err
	}
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, note.GameID)
	if err != nil {
		return models.Note{}, ErrForbidden
	}

	isAuthor := note.UserID == userID
	isGM := membership.IsGM

	if !isAuthor && !isGM {
		// Other players can only edit shared notes but cannot change visibility
		if note.Visibility != "shared" {
			return models.Note{}, ErrForbidden
		}
		delete(updates, "visibility")
	} else if isGM && !isAuthor {
		// GM can edit but not change visibility unless also the author
		delete(updates, "visibility")
	}

	// Extract version for optimistic locking
	var expectedVersion *int
	if v, ok := updates["version"]; ok {
		switch val := v.(type) {
		case int:
			expectedVersion = &val
		case float64:
			intVal := int(val)
			expectedVersion = &intVal
		}
		delete(updates, "version")
	}

	// Strip immutable fields
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	delete(updates, "game_id")
	delete(updates, "user_id")

	return s.repo.Update(id, updates, expectedVersion)
}

func (s *noteService) DeleteNote(id, userID uuid.UUID) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, note.GameID)
	if err != nil {
		return ErrForbidden
	}
	if note.UserID != userID && !membership.IsGM {
		return ErrForbidden
	}
	if err := s.repo.ClearNoteFromPins(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
