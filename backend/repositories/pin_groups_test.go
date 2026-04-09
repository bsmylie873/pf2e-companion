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

func TestPinGroupRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	id := uuid.New()
	group := &models.PinGroup{ID: id, GameID: uuid.New(), X: 0.5, Y: 0.5}

	// Insert
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "pin_groups"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	// Reload via First — GORM adds both the string condition and primary-key condition
	reloadRows := sqlmock.NewRows([]string{"id", "game_id"}).AddRow(id, group.GameID)
	mock.ExpectQuery(`SELECT \* FROM "pin_groups"`).
		WithArgs(id, id, 1).
		WillReturnRows(reloadRows)

	err := repo.Create(group)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinGroupRepository_Create_InsertError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	group := &models.PinGroup{ID: uuid.New(), GameID: uuid.New()}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "pin_groups"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(group)
	assert.Error(t, err)
}

func TestPinGroupRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id"}).AddRow(id, uuid.New())

	mock.ExpectQuery(`SELECT \* FROM "pin_groups"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	group, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, group.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinGroupRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "pin_groups"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestPinGroupRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id"}).
		AddRow(uuid.New(), gameID)

	mock.ExpectQuery(`SELECT \* FROM "pin_groups"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	groups, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinGroupRepository_FindByMapID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	mapID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "map_id"}).
		AddRow(uuid.New(), mapID)

	mock.ExpectQuery(`SELECT \* FROM "pin_groups"`).
		WithArgs(mapID).
		WillReturnRows(rows)

	groups, err := repo.FindByMapID(mapID)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinGroupRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "pin_groups"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinGroupRepository_CountMembers_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	groupID := uuid.New()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "session_pins"`).
		WithArgs(groupID).
		WillReturnRows(countRows)

	count, err := repo.CountMembers(groupID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPinGroupRepository_CountMembers_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPinGroupRepository(db)

	groupID := uuid.New()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "session_pins"`).
		WithArgs(groupID).
		WillReturnError(errors.New("db error"))

	_, err := repo.CountMembers(groupID)
	assert.Error(t, err)
}
