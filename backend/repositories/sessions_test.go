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

func TestSessionRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	id := uuid.New()
	s := &models.Session{ID: id, GameID: uuid.New(), Title: "Session 1"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "sessions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(s)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	s := &models.Session{ID: uuid.New(), GameID: uuid.New(), Title: "Session 1"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "sessions"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(s)
	assert.Error(t, err)
}

func TestSessionRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(uuid.New(), gameID, "Session 1")

	mock.ExpectQuery(`SELECT \* FROM "sessions"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	sessions, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByGameID_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	gameID := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "sessions"`).
		WithArgs(gameID).
		WillReturnError(errors.New("db error"))

	_, err := repo.FindByGameID(gameID)
	assert.Error(t, err)
}

func TestSessionRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(id, uuid.New(), "Session 1")

	mock.ExpectQuery(`SELECT \* FROM "sessions"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	session, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, session.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "sessions"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestSessionRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "sessions"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByGameIDPaginated_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	gameID := uuid.New()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sessions"`).
		WithArgs(gameID).
		WillReturnRows(countRows)

	dataRows := sqlmock.NewRows([]string{"id", "game_id", "title"}).
		AddRow(uuid.New(), gameID, "S1").
		AddRow(uuid.New(), gameID, "S2")
	mock.ExpectQuery(`SELECT \* FROM "sessions"`).
		WillReturnRows(dataRows)

	sessions, count, err := repo.FindByGameIDPaginated(gameID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Len(t, sessions, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSessionRepository_FindByGameIDPaginated_CountError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSessionRepository(db)

	gameID := uuid.New()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "sessions"`).
		WithArgs(gameID).
		WillReturnError(errors.New("db error"))

	_, _, err := repo.FindByGameIDPaginated(gameID, 0, 10)
	assert.Error(t, err)
}
