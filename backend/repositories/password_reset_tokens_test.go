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

func TestPasswordResetTokenRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	id := uuid.New()
	token := &models.PasswordResetToken{
		ID:        id,
		UserID:    uuid.New(),
		TokenHash: "resethash",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "password_reset_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetTokenRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	token := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "resethash",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "password_reset_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Create(token)
	assert.Error(t, err)
}

func TestPasswordResetTokenRepository_FindByTokenHash_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	id := uuid.New()
	hash := "resethash"
	rows := sqlmock.NewRows([]string{"id", "token_hash"}).AddRow(id, hash)

	mock.ExpectQuery(`SELECT \* FROM "password_reset_tokens"`).
		WithArgs(hash, 1).
		WillReturnRows(rows)

	token, err := repo.FindByTokenHash(hash)
	require.NoError(t, err)
	assert.Equal(t, id, token.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetTokenRepository_FindByTokenHash_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "password_reset_tokens"`).
		WithArgs("unknown", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.FindByTokenHash("unknown")
	assert.Error(t, err)
}

func TestPasswordResetTokenRepository_MarkUsed_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	hash := "resethash"
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "password_reset_tokens"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.MarkUsed(hash)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetTokenRepository_MarkUsed_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "password_reset_tokens"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.MarkUsed("resethash")
	assert.Error(t, err)
}

func TestPasswordResetTokenRepository_DeleteExpiredForUser_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	userID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "password_reset_tokens"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.DeleteExpiredForUser(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetTokenRepository_DeleteAllForUser_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPasswordResetTokenRepository(db)

	userID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "password_reset_tokens"`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 3))
	mock.ExpectCommit()

	err := repo.DeleteAllForUser(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
