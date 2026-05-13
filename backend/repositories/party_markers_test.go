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

// -- FindByGameID --

func TestPartyMarkerRepository_FindByGameID_Found(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPartyMarkerRepository(db)

	gameID := uuid.New()
	mapID := uuid.New()
	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "game_id", "map_id", "x", "y", "created_at", "updated_at"}).
		AddRow(id, gameID, mapID, 0.5, 0.3, now, now)

	mock.ExpectQuery(`SELECT \* FROM "party_markers"`).
		WithArgs(gameID, 1).
		WillReturnRows(rows)

	result, err := repo.FindByGameID(gameID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, gameID, result.GameID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPartyMarkerRepository_FindByGameID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPartyMarkerRepository(db)

	gameID := uuid.New()

	// Empty result set causes gorm First() to return ErrRecordNotFound.
	mock.ExpectQuery(`SELECT \* FROM "party_markers"`).
		WithArgs(gameID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	result, err := repo.FindByGameID(gameID)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPartyMarkerRepository_FindByGameID_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPartyMarkerRepository(db)

	gameID := uuid.New()
	dbErr := errors.New("connection refused")

	mock.ExpectQuery(`SELECT \* FROM "party_markers"`).
		WithArgs(gameID, 1).
		WillReturnError(dbErr)

	result, err := repo.FindByGameID(gameID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// -- Upsert --

func TestPartyMarkerRepository_Upsert_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPartyMarkerRepository(db)

	gameID := uuid.New()
	mapID := uuid.New()
	id := uuid.New()
	now := time.Now()

	marker := &models.PartyMarker{
		GameID: gameID,
		MapID:  mapID,
		X:      0.5,
		Y:      0.3,
	}

	// The raw INSERT ... ON CONFLICT exec.
	mock.ExpectExec(`INSERT INTO party_markers`).
		WithArgs(gameID, mapID, 0.5, 0.3).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Re-fetch via db.First(marker, "game_id = ?", marker.GameID).
	reloadRows := sqlmock.NewRows([]string{"id", "game_id", "map_id", "x", "y", "created_at", "updated_at"}).
		AddRow(id, gameID, mapID, 0.5, 0.3, now, now)
	mock.ExpectQuery(`SELECT \* FROM "party_markers"`).
		WithArgs(gameID, 1).
		WillReturnRows(reloadRows)

	err := repo.Upsert(marker)
	assert.NoError(t, err)
	assert.Equal(t, id, marker.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPartyMarkerRepository_Upsert_ExecError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPartyMarkerRepository(db)

	gameID := uuid.New()
	mapID := uuid.New()
	marker := &models.PartyMarker{
		GameID: gameID,
		MapID:  mapID,
		X:      0.5,
		Y:      0.3,
	}

	mock.ExpectExec(`INSERT INTO party_markers`).
		WithArgs(gameID, mapID, 0.5, 0.3).
		WillReturnError(errors.New("db error"))

	err := repo.Upsert(marker)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// -- Delete --

func TestPartyMarkerRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPartyMarkerRepository(db)

	gameID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "party_markers"`).
		WithArgs(gameID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(gameID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
