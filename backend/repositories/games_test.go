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

func TestGameRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()
	game := &models.Game{ID: id, Title: "Test Game"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "games"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(game)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	game := &models.Game{ID: uuid.New(), Title: "Test Game"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "games"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(game)
	assert.Error(t, err)
}

func TestGameRepository_FindAll_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id1 := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "title"}).
		AddRow(id1, "Game 1")

	mock.ExpectQuery(`SELECT \* FROM "games"`).WillReturnRows(rows)

	games, err := repo.FindAll()
	require.NoError(t, err)
	assert.Len(t, games, 1)
	assert.Equal(t, id1, games[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepository_FindAll_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "games"`).WillReturnError(errors.New("db error"))

	games, err := repo.FindAll()
	assert.Error(t, err)
	assert.Nil(t, games)
}

func TestGameRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "title"}).AddRow(id, "Test Game")

	mock.ExpectQuery(`SELECT \* FROM "games"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	game, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, game.ID)
	assert.Equal(t, "Test Game", game.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "games"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestGameRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "games"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepository_Delete_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "games"`).
		WithArgs(id).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Delete(id)
	assert.Error(t, err)
}

func TestGameRepository_FindByIDs_Empty(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewGameRepository(db)

	games, err := repo.FindByIDs([]uuid.UUID{})
	assert.NoError(t, err)
	assert.Empty(t, games)
}

func TestGameRepository_FindByIDs_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id1 := uuid.New()
	id2 := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "title"}).
		AddRow(id1, "Game 1").
		AddRow(id2, "Game 2")

	mock.ExpectQuery(`SELECT \* FROM "games"`).
		WillReturnRows(rows)

	games, err := repo.FindByIDs([]uuid.UUID{id1, id2})
	require.NoError(t, err)
	assert.Len(t, games, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepository_FindByIDsPaginated_Empty(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewGameRepository(db)

	games, count, err := repo.FindByIDsPaginated([]uuid.UUID{}, 0, 10)
	assert.NoError(t, err)
	assert.Empty(t, games)
	assert.Equal(t, int64(0), count)
}

func TestGameRepository_FindByIDsPaginated_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()

	// Count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "games"`).
		WillReturnRows(countRows)

	// Data query
	dataRows := sqlmock.NewRows([]string{"id", "title"}).AddRow(id, "Game 1")
	mock.ExpectQuery(`SELECT \* FROM "games"`).
		WillReturnRows(dataRows)

	games, count, err := repo.FindByIDsPaginated([]uuid.UUID{id}, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	assert.Len(t, games, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGameRepository_FindByIDsPaginated_CountError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGameRepository(db)

	id := uuid.New()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "games"`).
		WillReturnError(errors.New("db error"))

	_, _, err := repo.FindByIDsPaginated([]uuid.UUID{id}, 0, 10)
	assert.Error(t, err)
}
