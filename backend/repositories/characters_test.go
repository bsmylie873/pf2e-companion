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

func TestCharacterRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	id := uuid.New()
	char := &models.Character{ID: id, GameID: uuid.New(), Name: "Frodo"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "characters"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(char)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCharacterRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	char := &models.Character{ID: uuid.New(), GameID: uuid.New(), Name: "Frodo"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "characters"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(char)
	assert.Error(t, err)
}

func TestCharacterRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), gameID, "Frodo").
		AddRow(uuid.New(), gameID, "Sam")

	mock.ExpectQuery(`SELECT \* FROM "characters"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	chars, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, chars, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCharacterRepository_FindByGameID_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	gameID := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "characters"`).
		WithArgs(gameID).
		WillReturnError(errors.New("db error"))

	_, err := repo.FindByGameID(gameID)
	assert.Error(t, err)
}

func TestCharacterRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(id, uuid.New(), "Gandalf")

	mock.ExpectQuery(`SELECT \* FROM "characters"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	char, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, char.ID)
	assert.Equal(t, "Gandalf", char.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCharacterRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "characters"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestCharacterRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "characters"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCharacterRepository_Delete_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewCharacterRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "characters"`).
		WithArgs(id).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Delete(id)
	assert.Error(t, err)
}
