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

func TestPinRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	id := uuid.New()
	pin := &models.SessionPin{ID: id, GameID: uuid.New(), Label: "Pin 1", X: 0.5, Y: 0.5}

	// Insert
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "session_pins"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	// Reload via First — GORM adds both the string condition and primary-key condition
	reloadRows := sqlmock.NewRows([]string{"id", "label"}).AddRow(id, "Pin 1")
	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WithArgs(id, id, 1).
		WillReturnRows(reloadRows)

	err := repo.Create(pin)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_Create_InsertError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	pin := &models.SessionPin{ID: uuid.New(), GameID: uuid.New(), Label: "Pin 1"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "session_pins"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(pin)
	assert.Error(t, err)
}

func TestPinRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "label"}).
		AddRow(uuid.New(), gameID, "Pin A")

	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	pins, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "label"}).AddRow(id, "Pin A")

	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	pin, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, pin.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestPinRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "session_pins"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_FindByGroupID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	groupID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "group_id"}).
		AddRow(uuid.New(), groupID)

	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WithArgs(groupID).
		WillReturnRows(rows)

	pins, err := repo.FindByGroupID(groupID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_ClearGroupID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	pinID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "session_pins"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.ClearGroupID(pinID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_ClearGroupID_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	pinID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "session_pins"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.ClearGroupID(pinID)
	assert.Error(t, err)
}

func TestPinRepository_SetGroupID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	pinID := uuid.New()
	groupID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "session_pins"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SetGroupID(pinID, groupID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_FindStandaloneByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "label"}).
		AddRow(uuid.New(), gameID, "Standalone Pin")

	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WillReturnRows(rows)

	pins, err := repo.FindStandaloneByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_FindByMapID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	mapID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "map_id", "label"}).
		AddRow(uuid.New(), mapID, "Map Pin")

	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WithArgs(mapID).
		WillReturnRows(rows)

	pins, err := repo.FindByMapID(mapID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinRepository_FindStandaloneByMapID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinRepository(db)

	mapID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "map_id", "label"}).
		AddRow(uuid.New(), mapID, "Standalone Map Pin")

	mock.ExpectQuery(`SELECT \* FROM "session_pins"`).
		WillReturnRows(rows)

	pins, err := repo.FindStandaloneByMapID(mapID)
	require.NoError(t, err)
	assert.Len(t, pins, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
