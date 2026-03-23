package services

import (
	"github.com/google/uuid"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

type NoteService interface {
	CreateGameNote(gameID, userID uuid.UUID, note *models.Note) (models.Note, error)
	CreateUserNote(ownerID, callerID uuid.UUID, note *models.Note) (models.Note, error)
	ListGameNotes(gameID, userID uuid.UUID) ([]models.Note, error)
	ListUserNotes(ownerID, callerID uuid.UUID) ([]models.Note, error)
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

func (s *noteService) CreateGameNote(gameID, userID uuid.UUID, note *models.Note) (models.Note, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return models.Note{}, ErrForbidden
	}
	note.ID = uuid.Nil
	note.GameID = &gameID
	note.UserID = nil
	if err := s.repo.Create(note); err != nil {
		return models.Note{}, err
	}
	return *note, nil
}

func (s *noteService) CreateUserNote(ownerID, callerID uuid.UUID, note *models.Note) (models.Note, error) {
	if ownerID != callerID {
		return models.Note{}, ErrForbidden
	}
	note.ID = uuid.Nil
	note.UserID = &ownerID
	note.GameID = nil
	if err := s.repo.Create(note); err != nil {
		return models.Note{}, err
	}
	return *note, nil
}

func (s *noteService) ListGameNotes(gameID, userID uuid.UUID) ([]models.Note, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}
	return s.repo.FindByGameID(gameID)
}

func (s *noteService) ListUserNotes(ownerID, callerID uuid.UUID) ([]models.Note, error) {
	if ownerID != callerID {
		return nil, ErrForbidden
	}
	return s.repo.FindByUserID(ownerID)
}

func (s *noteService) GetNote(id, userID uuid.UUID) (models.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return models.Note{}, err
	}
	if note.UserID != nil {
		if *note.UserID != userID {
			return models.Note{}, ErrForbidden
		}
	} else if note.GameID != nil {
		if _, err := s.membershipRepo.FindByUserAndGameID(userID, *note.GameID); err != nil {
			return models.Note{}, ErrForbidden
		}
	}
	return note, nil
}

func (s *noteService) UpdateNote(id, userID uuid.UUID, updates map[string]interface{}) (models.Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return models.Note{}, err
	}
	if note.UserID != nil {
		if *note.UserID != userID {
			return models.Note{}, ErrForbidden
		}
	} else if note.GameID != nil {
		if _, err := s.membershipRepo.FindByUserAndGameID(userID, *note.GameID); err != nil {
			return models.Note{}, ErrForbidden
		}
	}
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	return s.repo.Update(id, updates)
}

func (s *noteService) DeleteNote(id, userID uuid.UUID) error {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if note.UserID != nil {
		if *note.UserID != userID {
			return ErrForbidden
		}
	} else if note.GameID != nil {
		if _, err := s.membershipRepo.FindByUserAndGameID(userID, *note.GameID); err != nil {
			return ErrForbidden
		}
	}
	return s.repo.Delete(id)
}
