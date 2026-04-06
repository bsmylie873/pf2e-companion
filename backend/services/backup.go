package services

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

// BackupService handles game export and import operations.
type BackupService interface {
	ExportGame(gameID, userID uuid.UUID) (*models.BackupFile, error)
	ExportSession(sessionID, userID uuid.UUID) (*models.BackupFile, error)
	ExportNote(noteID, userID uuid.UUID) (*models.BackupFile, error)
	ImportGame(gameID, userID uuid.UUID, mode string, backup *models.BackupFile) (*models.ImportSummary, error)
}

type backupService struct {
	db             *gorm.DB
	sessionRepo    repositories.SessionRepository
	noteRepo       repositories.NoteRepository
	membershipRepo repositories.MembershipRepository
}

// NewBackupService constructs a BackupService with the required dependencies.
func NewBackupService(
	db *gorm.DB,
	sessionRepo repositories.SessionRepository,
	noteRepo repositories.NoteRepository,
	membershipRepo repositories.MembershipRepository,
) BackupService {
	return &backupService{
		db:             db,
		sessionRepo:    sessionRepo,
		noteRepo:       noteRepo,
		membershipRepo: membershipRepo,
	}
}

// ExportGame exports all visible sessions and notes for a game.
func (s *backupService) ExportGame(gameID, userID uuid.UUID) (*models.BackupFile, error) {
	membership, err := s.membershipRepo.FindByUserAndGameID(userID, gameID)
	if err != nil {
		return nil, ErrForbidden
	}

	sessions, err := s.sessionRepo.FindByGameID(gameID)
	if err != nil {
		return nil, err
	}

	notes, err := s.noteRepo.FindByGameID(gameID, userID, membership.IsGM, repositories.NoteFilters{})
	if err != nil {
		return nil, err
	}

	return &models.BackupFile{
		SchemaVersion: "1",
		GameID:        gameID,
		ExportedAt:    time.Now().UTC(),
		Sessions:      sessions,
		Notes:         notes,
	}, nil
}

// ExportSession exports a single session and its associated notes.
func (s *backupService) ExportSession(sessionID, userID uuid.UUID) (*models.BackupFile, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, err
	}

	membership, err := s.membershipRepo.FindByUserAndGameID(userID, session.GameID)
	if err != nil {
		return nil, ErrForbidden
	}

	notes, err := s.noteRepo.FindByGameID(session.GameID, userID, membership.IsGM, repositories.NoteFilters{SessionID: &sessionID})
	if err != nil {
		return nil, err
	}

	return &models.BackupFile{
		SchemaVersion: "1",
		GameID:        session.GameID,
		ExportedAt:    time.Now().UTC(),
		Sessions:      []models.Session{session},
		Notes:         notes,
	}, nil
}

// ExportNote exports a single note.
func (s *backupService) ExportNote(noteID, userID uuid.UUID) (*models.BackupFile, error) {
	note, err := s.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, err
	}

	membership, err := s.membershipRepo.FindByUserAndGameID(userID, note.GameID)
	if err != nil {
		return nil, ErrForbidden
	}

	// Visibility check: GM sees all; non-GM sees visible, editable, or own notes.
	if !membership.IsGM {
		if note.Visibility != "visible" && note.Visibility != "editable" && note.UserID != userID {
			return nil, ErrForbidden
		}
	}

	return &models.BackupFile{
		SchemaVersion: "1",
		GameID:        note.GameID,
		ExportedAt:    time.Now().UTC(),
		Sessions:      []models.Session{},
		Notes:         []models.Note{note},
	}, nil
}

// ImportGame imports sessions and notes from a backup file into an existing game.
func (s *backupService) ImportGame(gameID, userID uuid.UUID, mode string, backup *models.BackupFile) (*models.ImportSummary, error) {
	if _, err := s.membershipRepo.FindByUserAndGameID(userID, gameID); err != nil {
		return nil, ErrForbidden
	}

	// Build set of valid member user IDs for this game.
	memberships, err := s.membershipRepo.FindByGameID(gameID)
	if err != nil {
		return nil, err
	}
	validUserIDs := make(map[uuid.UUID]bool)
	for _, m := range memberships {
		validUserIDs[m.UserID] = true
	}

	// Build set of session IDs included in this import file.
	importedSessionIDs := make(map[uuid.UUID]bool)
	for _, sess := range backup.Sessions {
		importedSessionIDs[sess.ID] = true
	}

	summary := &models.ImportSummary{}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Process sessions first (order matters for session_id references in notes).
		for _, sess := range backup.Sessions {
			sess.FolderID = nil
			sess.GameID = gameID

			var existing models.Session
			existsErr := tx.First(&existing, "id = ?", sess.ID).Error
			if existsErr == nil {
				// Record exists.
				if mode == "merge" {
					summary.SessionsSkipped++
					continue
				}
				// Overwrite mode.
				if err := tx.Save(&sess).Error; err != nil {
					return err
				}
				summary.SessionsOverwritten++
			} else if existsErr == gorm.ErrRecordNotFound {
				if err := tx.Create(&sess).Error; err != nil {
					return err
				}
				summary.SessionsCreated++
			} else {
				return existsErr
			}
		}

		// Process notes.
		for _, note := range backup.Notes {
			note.FolderID = nil
			note.GameID = gameID

			// Resolve session_id: keep if in import set or already exists in DB, else clear.
			if note.SessionID != nil {
				if !importedSessionIDs[*note.SessionID] {
					var existingSession models.Session
					if err := tx.First(&existingSession, "id = ?", *note.SessionID).Error; err != nil {
						note.SessionID = nil
					}
				}
			}

			// Substitute user_id if not a valid game member.
			if !validUserIDs[note.UserID] {
				note.UserID = userID
			}

			// Preserve visibility as-is.

			var existing models.Note
			existsErr := tx.First(&existing, "id = ?", note.ID).Error
			if existsErr == nil {
				if mode == "merge" {
					summary.NotesSkipped++
					continue
				}
				if err := tx.Save(&note).Error; err != nil {
					return err
				}
				summary.NotesOverwritten++
			} else if existsErr == gorm.ErrRecordNotFound {
				if err := tx.Create(&note).Error; err != nil {
					return err
				}
				summary.NotesCreated++
			} else {
				return existsErr
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return summary, nil
}
