package services

import (
	"fmt"

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
	folderRepo     repositories.FolderRepository
}

func NewNoteService(repo repositories.NoteRepository, membershipRepo repositories.MembershipRepository, folderRepo repositories.FolderRepository) NoteService {
	return &noteService{repo: repo, membershipRepo: membershipRepo, folderRepo: folderRepo}
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
	if note.Visibility == "visible" || note.Visibility == "editable" {
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
		// Non-author, non-GM players
		if note.Visibility == "editable" {
			// Can edit content, cannot change visibility
			delete(updates, "visibility")
		} else {
			// private or visible — no edit access
			return models.Note{}, ErrForbidden
		}
	} else if isGM && !isAuthor {
		// GM can edit content but not change visibility
		delete(updates, "visibility")
	}

	// Strip immutable fields
	delete(updates, "id")
	delete(updates, "created_at")
	delete(updates, "updated_at")
	delete(updates, "game_id")
	delete(updates, "user_id")

	// Folder-visibility validation
	if newVis, ok := updates["visibility"]; ok {
		visStr, _ := newVis.(string)
		if visStr == "editable" || visStr == "gm-only" {
			if note.FolderID != nil {
				folder, ferr := s.folderRepo.FindByID(*note.FolderID)
				if ferr == nil && folder.Visibility == "private" {
					return models.Note{}, fmt.Errorf("%w: cannot change visibility while note is in a private folder", ErrValidation)
				}
			}
		}
	}
	if newFolderID, ok := updates["folder_id"]; ok && newFolderID != nil {
		var fid uuid.UUID
		switch v := newFolderID.(type) {
		case string:
			fid, _ = uuid.Parse(v)
		case uuid.UUID:
			fid = v
		}
		if fid != uuid.Nil {
			folder, ferr := s.folderRepo.FindByID(fid)
			if ferr == nil && folder.Visibility == "private" && note.Visibility != "private" {
				return models.Note{}, fmt.Errorf("%w: only private notes can be placed in a private folder", ErrValidation)
			}
		}
	}

	return s.repo.Update(id, updates)
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
