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

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	user := &models.User{ID: id, Username: "alice", Email: "alice@example.com"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DBError(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{ID: uuid.New(), Username: "alice", Email: "alice@example.com"}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(errors.New("duplicate key"))
	mock.ExpectRollback()

	err := repo.Create(user)
	assert.Error(t, err)
}

func TestUserRepository_FindAll_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(id, "alice", "alice@example.com")

	mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnRows(rows)

	users, err := repo.FindAll()
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "alice", users[0].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(id, "alice", "alice@example.com")

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(id, 1).
		WillReturnRows(rows)

	user, err := repo.FindByID(id)
	require.NoError(t, err)
	assert.Equal(t, id, user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(id, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}))

	_, err := repo.FindByID(id)
	assert.Error(t, err)
}

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "users"`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Delete(id)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	email := "alice@example.com"
	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(id, "alice", email)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs(email, 1).
		WillReturnRows(rows)

	user, err := repo.FindByEmail(email)
	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs("nobody@example.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email"}))

	_, err := repo.FindByEmail("nobody@example.com")
	assert.Error(t, err)
}

func TestUserRepository_FindByUsername_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	id := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "username", "email"}).
		AddRow(id, "alice", "alice@example.com")

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs("alice", 1).
		WillReturnRows(rows)

	user, err := repo.FindByUsername("alice")
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WithArgs("ghost", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username"}))

	_, err := repo.FindByUsername("ghost")
	assert.Error(t, err)
}
