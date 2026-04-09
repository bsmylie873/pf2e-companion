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

func TestFolderRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := uuid.New()
	folder := &models.Folder{ID: id, GameID: uuid.New(), Name: "Session Folder", FolderType: "session"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "folders"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(folder)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(id, uuid.New(), "My Folder")

	mock.ExpectQuery(`SELECT \* FROM "folders"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	folder, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, folder.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "folders"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestFolderRepository_FindSessionFolders_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name", "folder_type"}).
		AddRow(uuid.New(), gameID, "Session Folder", "session")

	mock.ExpectQuery(`SELECT \* FROM "folders"`).
		WillReturnRows(rows)

	folders, err := repo.FindSessionFolders(gameID)
	require.NoError(t, err)
	assert.Len(t, folders, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_FindNoteFolders_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	gameID := uuid.New()
	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name", "folder_type"}).
		AddRow(uuid.New(), gameID, "Note Folder", "note")

	mock.ExpectQuery(`SELECT \* FROM "folders"`).
		WillReturnRows(rows)

	folders, err := repo.FindNoteFolders(gameID, userID)
	require.NoError(t, err)
	assert.Len(t, folders, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "folders"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_FindAllByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), gameID, "Folder A").
		AddRow(uuid.New(), gameID, "Folder B")

	mock.ExpectQuery(`SELECT \* FROM "folders"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	folders, err := repo.FindAllByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, folders, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_BatchUpdatePositions_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	id1 := uuid.New()
	id2 := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "folders"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`UPDATE "folders"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.BatchUpdatePositions([]uuid.UUID{id1, id2}, []int{0, 1})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_BatchUpdatePositions_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	id1 := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "folders"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.BatchUpdatePositions([]uuid.UUID{id1}, []int{0})
	assert.Error(t, err)
}

func TestFolderRepository_MaxPosition_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"coalesce"}).AddRow(5)

	mock.ExpectQuery(`SELECT COALESCE`).
		WillReturnRows(rows)

	pos, err := repo.MaxPosition(gameID, "session", nil)
	require.NoError(t, err)
	assert.Equal(t, 5, pos)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_MaxPosition_WithUserID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	gameID := uuid.New()
	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"coalesce"}).AddRow(2)

	mock.ExpectQuery(`SELECT COALESCE`).
		WillReturnRows(rows)

	pos, err := repo.MaxPosition(gameID, "note", &userID)
	require.NoError(t, err)
	assert.Equal(t, 2, pos)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFolderRepository_MaxPosition_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewFolderRepository(db)

	gameID := uuid.New()
	mock.ExpectQuery(`SELECT COALESCE`).
		WillReturnError(errors.New("db error"))

	_, err := repo.MaxPosition(gameID, "session", nil)
	assert.Error(t, err)
}
