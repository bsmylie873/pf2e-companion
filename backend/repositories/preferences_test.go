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

func TestPreferenceRepository_FindByUserID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	userID := uuid.New()
	rows := sqlmock.NewRows([]string{"user_id"}).AddRow(userID)

	mock.ExpectQuery(`SELECT \* FROM "user_preferences"`).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	pref, err := repo.FindByUserID(userID)
	require.NoError(t, err)
	assert.Equal(t, userID, pref.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPreferenceRepository_FindByUserID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	userID := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "user_preferences"`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	_, err := repo.FindByUserID(userID)
	assert.Error(t, err)
}

func TestPreferenceRepository_Upsert_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	userID := uuid.New()
	pref := &models.UserPreference{UserID: userID}

	// Save does an upsert; GORM checks if record exists and updates or inserts.
	// With a set primary key GORM issues an UPDATE first.
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_preferences"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Upsert(pref)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPreferenceRepository_Upsert_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	pref := &models.UserPreference{UserID: uuid.New()}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_preferences"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.Upsert(pref)
	assert.Error(t, err)
}

func TestPreferenceRepository_ClearDefaultGameForGame_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	gameID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_preferences"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.ClearDefaultGameForGame(gameID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPreferenceRepository_ClearDefaultGameForGame_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	gameID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_preferences"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	err := repo.ClearDefaultGameForGame(gameID)
	assert.Error(t, err)
}

func TestPreferenceRepository_ClearDefaultGameForMembership_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPreferenceRepository(db)

	userID := uuid.New()
	gameID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_preferences"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.ClearDefaultGameForMembership(userID, gameID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
