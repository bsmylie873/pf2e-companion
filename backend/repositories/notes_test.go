package repositories

import (
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pf2e-companion/backend/models"
)

func TestNoteRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	id := uuid.New()
	note := &models.Note{ID: id, GameID: uuid.New(), UserID: uuid.New(), Title: "Test Note"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "notes"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(note)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	note := &models.Note{ID: uuid.New(), GameID: uuid.New(), UserID: uuid.New(), Title: "Test Note"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "notes"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(note)
	assert.Error(t, err)
}

func TestNoteRepository_FindByGameID_AsGM_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	gameID := uuid.New()
	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(uuid.New(), gameID, "Note 1")

	mock.ExpectQuery(`SELECT \* FROM "notes"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	notes, err := repo.FindByGameID(gameID, userID, true, NoteFilters{})
	require.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_FindByGameID_AsPlayer_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	gameID := uuid.New()
	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(uuid.New(), gameID, "Note 1")

	mock.ExpectQuery(`SELECT \* FROM "notes"`).
		WillReturnRows(rows)

	notes, err := repo.FindByGameID(gameID, userID, false, NoteFilters{})
	require.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_FindByGameID_WithSessionFilter(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	gameID := uuid.New()
	userID := uuid.New()
	sessionID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(uuid.New(), gameID, "Note 1")

	mock.ExpectQuery(`SELECT \* FROM "notes"`).
		WillReturnRows(rows)

	notes, err := repo.FindByGameID(gameID, userID, true, NoteFilters{SessionID: &sessionID})
	require.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(id, uuid.New(), "Test Note")

	mock.ExpectQuery(`SELECT \* FROM "notes"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	note, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, note.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "notes"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestNoteRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "notes"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_ClearNoteFromPins_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	noteID := uuid.New()

	// GORM UPDATE generates: SET "note_id"=$1,"updated_at"=$2 WHERE note_id = $3
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "session_pins"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), noteID).
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.ClearNoteFromPins(noteID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoteRepository_ClearNoteFromPins_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	noteID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "session_pins"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.ClearNoteFromPins(noteID)
	assert.Error(t, err)
}

func TestNoteRepository_FindByGameIDPaginated_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoteRepository(db)

	gameID := uuid.New()
	userID := uuid.New()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "notes"`).
		WillReturnRows(countRows)

	dataRows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(uuid.New(), gameID, "Note 1")
	mock.ExpectQuery(`SELECT \* FROM "notes"`).
		WillReturnRows(dataRows)

	notes, count, err := repo.FindByGameIDPaginated(gameID, userID, true, NoteFilters{}, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Len(t, notes, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
