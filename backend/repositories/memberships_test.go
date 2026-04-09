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

func TestMembershipRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	id := uuid.New()
	m := &models.GameMembership{ID: id, GameID: uuid.New(), UserID: uuid.New(), IsGM: true}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "game_memberships"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(m)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	m := &models.GameMembership{ID: uuid.New(), GameID: uuid.New(), UserID: uuid.New()}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "game_memberships"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(m)
	assert.Error(t, err)
}

func TestMembershipRepository_FindByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	gameID := uuid.New()
	memberID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id"}).
		AddRow(memberID, gameID)

	mock.ExpectQuery(`SELECT \* FROM "game_memberships"`).
		WithArgs(gameID).
		WillReturnRows(rows)

	memberships, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	assert.Len(t, memberships, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "user_id"}).
		AddRow(id, uuid.New(), uuid.New())

	mock.ExpectQuery(`SELECT \* FROM "game_memberships"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	membership, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, membership.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "game_memberships"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestMembershipRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "game_memberships"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipRepository_FindByUserID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "user_id"}).
		AddRow(uuid.New(), userID)

	mock.ExpectQuery(`SELECT \* FROM "game_memberships"`).
		WithArgs(userID).
		WillReturnRows(rows)

	memberships, err := repo.FindByUserID(userID)
	require.NoError(t, err)
	assert.Len(t, memberships, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipRepository_FindByUserAndGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	userID := uuid.New()
	gameID := uuid.New()
	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "user_id", "game_id"}).
		AddRow(id, userID, gameID)

	mock.ExpectQuery(`SELECT \* FROM "game_memberships"`).
		WithArgs(userID, gameID, 1).
		WillReturnRows(rows)

	membership, err := repo.FindByUserAndGameID(userID, gameID)
	require.NoError(t, err)
	assert.Equal(t, id, membership.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMembershipRepository_FindByUserAndGameID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewMembershipRepository(db)

	userID := uuid.New()
	gameID := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "game_memberships"`).
		WithArgs(userID, gameID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByUserAndGameID(userID, gameID)
	assert.Error(t, err)
}
