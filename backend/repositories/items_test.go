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

func TestItemRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	id := uuid.New()
	item := &models.Item{ID: id, GameID: uuid.New(), Name: "Longsword"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "items"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(item)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	item := &models.Item{ID: uuid.New(), GameID: uuid.New(), Name: "Longsword"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "items"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(item)
	assert.Error(t, err)
}

func TestItemRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	gameID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), gameID, "Longsword").
		AddRow(uuid.New(), gameID, "Shield")

	mock.ExpectQuery(`SELECT \* FROM "items"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	items, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemRepository_FindByCharacterID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	charID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(uuid.New(), uuid.New(), "Dagger")

	mock.ExpectQuery(`SELECT \* FROM "items"`).
		WithArgs(charID).
		WillReturnRows(rows)

	items, err := repo.FindByCharacterID(charID)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemRepository_FindByCharacterID_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	charID := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "items"`).
		WithArgs(charID).
		WillReturnError(errors.New("db error"))

	_, err := repo.FindByCharacterID(charID)
	assert.Error(t, err)
}

func TestItemRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "name"}).
		AddRow(id, uuid.New(), "Shortsword")

	mock.ExpectQuery(`SELECT \* FROM "items"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	item, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, item.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "items"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestItemRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "items"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItemRepository_Delete_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewItemRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "items"`).
		WithArgs(id).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Delete(id)
	assert.Error(t, err)
}
