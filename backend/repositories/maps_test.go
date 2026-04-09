package repositories

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pf2e-companion/backend/models"
)

func TestMapRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	m := &models.GameMap{ID: id, GameID: uuid.New(), Name: "World Map"}

	// Insert
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "maps"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	// Reload via First — GORM adds both the string condition and primary-key condition
	reloadRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(id, "World Map")
	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WithArgs(id, id, 1).
		WillReturnRows(reloadRows)

	err := repo.Create(m)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_Create_InsertError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	m := &models.GameMap{ID: uuid.New(), GameID: uuid.New(), Name: "World Map"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "maps"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(m)
	assert.Error(t, err)
}

func TestMapRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(id, "World Map")

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	m, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, m.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestMapRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), gameID, "Map A")

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	maps, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, maps, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_FindActiveByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), gameID, "Active Map")

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WillReturnRows(rows)

	maps, err := repo.FindActiveByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, maps, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_FindArchivedByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), gameID, "Archived Map")

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WillReturnRows(rows)

	maps, err := repo.FindArchivedByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, maps, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_FindByGameIDAndName_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	gameID := uuid.New()
	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(id, gameID, "World Map")

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WillReturnRows(rows)

	m, err := repo.FindByGameIDAndName(gameID, "World Map")
	require.NoError(t, err)
	assert.Equal(t, id, m.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_Archive_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "maps"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Archive(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_Archive_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "maps"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Archive(id)
	assert.Error(t, err)
}

func TestMapRepository_HardDelete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "maps"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.HardDelete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_HardDelete_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "maps"`).
		WithArgs(id).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.HardDelete(id)
	assert.Error(t, err)
}

func TestMapRepository_FindExpiredArchived_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	cutoff := time.Now().Add(-24 * time.Hour)
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(uuid.New(), "Expired Map")

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WillReturnRows(rows)

	maps, err := repo.FindExpiredArchived(cutoff)
	require.NoError(t, err)
	assert.Len(t, maps, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMapRepository_FindExpiredArchived_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMapRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "maps"`).
		WillReturnError(errors.New("db error"))

	_, err := repo.FindExpiredArchived(time.Now())
	assert.Error(t, err)
}
