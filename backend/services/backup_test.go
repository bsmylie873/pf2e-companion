package services

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

func setupBackupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db, mock
}

func TestBackupService_ExportGame_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	sessions := []models.Session{{ID: uuid.New(), Title: "S1"}}
	notes := []models.Note{{ID: uuid.New(), Title: "N1"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("FindByGameID", gameID).Return(sessions, nil)
	mockNoteRepo.On("FindByGameID", gameID, userID, true, repositories.NoteFilters{}).Return(notes, nil)

	result, err := svc.ExportGame(gameID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1", result.SchemaVersion)
	assert.Equal(t, gameID, result.GameID)
	assert.Len(t, result.Sessions, 1)
	assert.Len(t, result.Notes, 1)
	mockMemberRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestBackupService_ExportGame_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ExportGame(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ExportSession_Success(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	gameID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID, Title: "Session 1"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	notes := []models.Note{}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("FindByGameID", gameID, userID, false, mock.AnythingOfType("repositories.NoteFilters")).Return(notes, nil)

	result, err := svc.ExportSession(sessionID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Sessions, 1)
	assert.Equal(t, sessionID, result.Sessions[0].ID)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestBackupService_ExportNote_Success_OwnNote(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: userID, Visibility: "private"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.ExportNote(noteID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Notes, 1)
	assert.Equal(t, noteID, result.Notes[0].ID)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ExportNote_Forbidden_PrivateOtherUser(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New(), Visibility: "private"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.ExportNote(noteID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ImportGame_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ImportGame(gameID, userID, "merge", &models.BackupFile{})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ExportGame_SessionsError(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("FindByGameID", gameID).Return([]models.Session{}, assert.AnError)

	_, err := svc.ExportGame(gameID, userID)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
}

func TestBackupService_ExportGame_NotesError(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	sessions := []models.Session{{ID: uuid.New(), Title: "S1"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockSessionRepo.On("FindByGameID", gameID).Return(sessions, nil)
	mockNoteRepo.On("FindByGameID", gameID, userID, true, repositories.NoteFilters{}).Return([]models.Note{}, assert.AnError)

	_, err := svc.ExportGame(gameID, userID)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestBackupService_ExportSession_SessionNotFound(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	sessionID := uuid.New()

	mockSessionRepo.On("FindByID", sessionID).Return(models.Session{}, gorm.ErrRecordNotFound)

	_, err := svc.ExportSession(sessionID, userID)

	assert.Error(t, err)
	mockSessionRepo.AssertExpectations(t)
}

func TestBackupService_ExportSession_Forbidden(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	gameID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID, Title: "Session 1"}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ExportSession(sessionID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ExportSession_NotesError(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	gameID := uuid.New()
	session := models.Session{ID: sessionID, GameID: gameID, Title: "Session 1"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockSessionRepo.On("FindByID", sessionID).Return(session, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("FindByGameID", gameID, userID, false, mock.AnythingOfType("repositories.NoteFilters")).Return([]models.Note{}, assert.AnError)

	_, err := svc.ExportSession(sessionID, userID)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrForbidden)
	mockSessionRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockNoteRepo.AssertExpectations(t)
}

func TestBackupService_ExportNote_NotFound(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	noteID := uuid.New()

	mockNoteRepo.On("FindByID", noteID).Return(models.Note{}, gorm.ErrRecordNotFound)

	_, err := svc.ExportNote(noteID, userID)

	assert.Error(t, err)
	mockNoteRepo.AssertExpectations(t)
}

func TestBackupService_ExportNote_Forbidden_MembershipError(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New(), Visibility: "private"}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, assert.AnError)

	_, err := svc.ExportNote(noteID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ExportNote_GM_CanExportPrivate(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	// Note belongs to a different user and is private — GM should still be able to export it
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New(), Visibility: "private"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.ExportNote(noteID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Notes, 1)
	assert.Equal(t, noteID, result.Notes[0].ID)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestBackupService_ImportGame_MembershipsError(t *testing.T) {
	mockSessionRepo := &mocks.MockSessionRepository{}
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	db, _ := setupBackupTestDB(t)
	svc := NewBackupService(db, mockSessionRepo, mockNoteRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMemberRepo.On("FindByGameID", gameID).Return([]models.GameMembership{}, assert.AnError)

	_, err := svc.ImportGame(gameID, userID, "merge", &models.BackupFile{})

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}
