package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestUserService_ListUsers_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	users := []models.User{
		{ID: uuid.New(), Username: "alice"},
		{ID: uuid.New(), Username: "bob"},
	}

	mockUserRepo.On("FindAll").Return(users, nil)

	result, err := svc.ListUsers()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "alice", result[0].Username)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_ListUsers_Error(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	mockUserRepo.On("FindAll").Return([]models.User{}, errors.New("db error"))

	_, err := svc.ListUsers()

	assert.Error(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUser_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()
	user := models.User{ID: userID, Username: "alice", Email: "alice@example.com"}

	mockUserRepo.On("FindByID", userID).Return(user, nil)

	resp, err := svc.GetUser(userID)

	assert.NoError(t, err)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, "alice@example.com", resp.Email)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()

	mockUserRepo.On("FindByID", userID).Return(models.User{}, gorm.ErrRecordNotFound)

	_, err := svc.GetUser(userID)

	assert.Error(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()
	user := models.User{ID: userID, Username: "alice"}
	updated := models.User{ID: userID, Username: "alice-updated"}
	updates := map[string]interface{}{"username": "alice-updated"}

	mockUserRepo.On("FindByID", userID).Return(user, nil)
	mockUserRepo.On("Update", userID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	resp, err := svc.UpdateUser(userID, updates, userID)

	assert.NoError(t, err)
	assert.Equal(t, "alice-updated", resp.Username)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_Forbidden(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()
	callerID := uuid.New()

	_, err := svc.UpdateUser(userID, map[string]interface{}{"username": "x"}, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUserService_UpdateUser_WithPasswordChange(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()
	user := models.User{ID: userID, Username: "alice"}
	updated := models.User{ID: userID, Username: "alice"}
	updates := map[string]interface{}{"password": "newpassword123"}

	mockUserRepo.On("FindByID", userID).Return(user, nil)
	mockUserRepo.On("Update", userID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)
	mockAuthSvc.On("InvalidateAllSessions", userID).Return(nil)

	resp, err := svc.UpdateUser(userID, updates, userID)

	assert.NoError(t, err)
	assert.Equal(t, "alice", resp.Username)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()
	user := models.User{ID: userID, Username: "alice"}

	mockUserRepo.On("FindByID", userID).Return(user, nil)
	mockUserRepo.On("Delete", userID).Return(nil)

	err := svc.DeleteUser(userID, userID)

	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_Forbidden(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()
	callerID := uuid.New()

	err := svc.DeleteUser(userID, callerID)

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockAuthSvc := &mocks.MockAuthService{}
	svc := NewUserService(mockUserRepo, mockAuthSvc)

	userID := uuid.New()

	mockUserRepo.On("FindByID", userID).Return(models.User{}, gorm.ErrRecordNotFound)

	err := svc.DeleteUser(userID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockUserRepo.AssertExpectations(t)
}
