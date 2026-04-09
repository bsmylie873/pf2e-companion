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

func TestInviteTokenRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	id := uuid.New()
	token := &models.InviteToken{
		ID:        id,
		GameID:    uuid.New(),
		CreatedBy: uuid.New(),
		TokenHash: "invitehash",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "invite_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInviteTokenRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	token := &models.InviteToken{
		ID:        uuid.New(),
		GameID:    uuid.New(),
		CreatedBy: uuid.New(),
		TokenHash: "invitehash",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "invite_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(token)
	assert.Error(t, err)
}

func TestInviteTokenRepository_FindActiveByGameID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	gameID := uuid.New()
	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "game_id", "token_hash"}).
		AddRow(id, gameID, "invitehash")

	mock.ExpectQuery(`SELECT \* FROM "invite_tokens"`).
		WillReturnRows(rows)

	token, err := repo.FindActiveByGameID(gameID)
	require.NoError(t, err)
	assert.Equal(t, id, token.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInviteTokenRepository_FindActiveByGameID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	gameID := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "invite_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindActiveByGameID(gameID)
	assert.Error(t, err)
}

func TestInviteTokenRepository_FindByTokenHash_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	id := uuid.New()
	hash := "invitehash"
	rows := sqlmock.NewRows([]string{"id", "token_hash"}).AddRow(id, hash)

	mock.ExpectQuery(`SELECT \* FROM "invite_tokens"`).
		WithArgs(hash, 1).
		WillReturnRows(rows)

	token, err := repo.FindByTokenHash(hash)
	require.NoError(t, err)
	assert.Equal(t, id, token.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInviteTokenRepository_FindByTokenHash_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "invite_tokens"`).
		WithArgs("unknown", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByTokenHash("unknown")
	assert.Error(t, err)
}

func TestInviteTokenRepository_RevokeAllForGame_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	gameID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "invite_tokens"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.RevokeAllForGame(gameID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInviteTokenRepository_RevokeAllForGame_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	gameID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "invite_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.RevokeAllForGame(gameID)
	assert.Error(t, err)
}

func TestInviteTokenRepository_RevokeByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "invite_tokens"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.RevokeByID(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInviteTokenRepository_RevokeByID_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewInviteTokenRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "invite_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.RevokeByID(id)
	assert.Error(t, err)
}

// Ensure time import is used (used in token structs above)
var _ = time.Now
