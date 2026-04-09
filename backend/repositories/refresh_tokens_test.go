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

func TestRefreshTokenRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	id := uuid.New()
	token := &models.RefreshToken{
		ID:        id,
		UserID:    uuid.New(),
		TokenHash: "hashval",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	token := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "hashval",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "refresh_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(token)
	assert.Error(t, err)
}

func TestRefreshTokenRepository_FindByTokenHash_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	id := uuid.New()
	hash := "myhash"
	rows := sqlmock.NewRows([]string{"id", "token_hash"}).AddRow(id, hash)

	mock.ExpectQuery(`SELECT \* FROM "refresh_tokens"`).
		WithArgs(hash, 1).
		WillReturnRows(rows)

	token, err := repo.FindByTokenHash(hash)
	require.NoError(t, err)
	assert.Equal(t, id, token.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_FindByTokenHash_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "refresh_tokens"`).
		WithArgs("badsig", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByTokenHash("badsig")
	assert.Error(t, err)
}

func TestRefreshTokenRepository_DeleteByTokenHash_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	hash := "myhash"
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens"`).
		WithArgs(hash).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteByTokenHash(hash)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_DeleteByTokenHash_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.DeleteByTokenHash("hashval")
	assert.Error(t, err)
}

func TestRefreshTokenRepository_DeleteExpiredForUser_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	userID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens"`).
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.DeleteExpiredForUser(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_DeleteAllForUser_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	userID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens"`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 5))
	mock.ExpectCommit()

	err := repo.DeleteAllForUser(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_DeleteAllForUser_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)

	userID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens"`).
		WithArgs(userID).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.DeleteAllForUser(userID)
	assert.Error(t, err)
}
