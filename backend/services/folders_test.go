package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestFolderService_CreateFolder_Success_NoteFolder(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("MaxPosition", gameID, "note", mock.Anything).Return(0, nil)
	mockFolderRepo.On("Create", mock.AnythingOfType("*models.Folder")).Return(nil)

	result, err := svc.CreateFolder(gameID, userID, "My Notes", "note", "private")

	assert.NoError(t, err)
	assert.Equal(t, "My Notes", result.Name)
	assert.Equal(t, 1, result.Position)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_SessionFolder_GM(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("MaxPosition", gameID, "session", mock.Anything).Return(0, nil)
	mockFolderRepo.On("Create", mock.AnythingOfType("*models.Folder")).Return(nil)

	result, err := svc.CreateFolder(gameID, userID, "Sessions", "session", "game-wide")

	assert.NoError(t, err)
	assert.Equal(t, "Sessions", result.Name)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_SessionFolder_NonGM_Fails(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateFolder(gameID, userID, "Sessions", "session", "game-wide")

	assert.ErrorIs(t, err, ErrSessionFoldersReadOnly)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_Forbidden(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateFolder(gameID, userID, "Name", "note", "private")

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_InvalidType(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateFolder(gameID, userID, "Name", "invalid_type", "private")

	assert.ErrorIs(t, err, ErrValidation)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_InvalidVisibility(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateFolder(gameID, userID, "Name", "note", "editable")

	assert.ErrorIs(t, err, ErrValidation)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_EmptyName(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateFolder(gameID, userID, "", "note", "private")

	assert.ErrorIs(t, err, ErrValidation)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_TooLongName(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	longName := strings.Repeat("a", 101)
	_, err := svc.CreateFolder(gameID, userID, longName, "note", "private")

	assert.ErrorIs(t, err, ErrValidation)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_MaxPositionError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	maxPosErr := errors.New("db error")

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("MaxPosition", gameID, "note", mock.Anything).Return(0, maxPosErr)

	_, err := svc.CreateFolder(gameID, userID, "Valid Name", "note", "private")

	assert.ErrorIs(t, err, maxPosErr)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_DuplicateConflict(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	dupErr := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("MaxPosition", gameID, "note", mock.Anything).Return(0, nil)
	mockFolderRepo.On("Create", mock.AnythingOfType("*models.Folder")).Return(dupErr)

	_, err := svc.CreateFolder(gameID, userID, "Valid Name", "note", "private")

	assert.ErrorIs(t, err, ErrConflict)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_CreateFolder_CreateOtherError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	createErr := errors.New("generic db error")

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("MaxPosition", gameID, "note", mock.Anything).Return(0, nil)
	mockFolderRepo.On("Create", mock.AnythingOfType("*models.Folder")).Return(createErr)

	_, err := svc.CreateFolder(gameID, userID, "Valid Name", "note", "private")

	assert.ErrorIs(t, err, createErr)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ListFolders_SessionFolders(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	folders := []models.Folder{{ID: uuid.New(), Name: "Sessions", Visibility: "game-wide"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindSessionFolders", gameID).Return(folders, nil)

	result, err := svc.ListFolders(gameID, userID, "session")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ListFolders_NoteFolders_FiltersPrivate(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	// One public folder, one private folder owned by another user
	folders := []models.Folder{
		{ID: uuid.New(), Name: "Public", Visibility: "game-wide"},
		{ID: uuid.New(), Name: "Private Other", Visibility: "private", UserID: &otherUserID},
	}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindNoteFolders", gameID, userID).Return(folders, nil)

	result, err := svc.ListFolders(gameID, userID, "note")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Public", result[0].Name)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ListFolders_Forbidden(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListFolders(gameID, userID, "note")

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ListFolders_RepoError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	repoErr := errors.New("db error")

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindNoteFolders", gameID, userID).Return([]models.Folder{}, repoErr)

	_, err := svc.ListFolders(gameID, userID, "note")

	assert.ErrorIs(t, err, repoErr)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ListFolders_PrivateNilUserID_Filtered(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	// Private folder with nil UserID should be filtered out
	folders := []models.Folder{
		{ID: uuid.New(), Name: "No Owner Private", Visibility: "private", UserID: nil},
		{ID: uuid.New(), Name: "Public", Visibility: "game-wide"},
	}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindNoteFolders", gameID, userID).Return(folders, nil)

	result, err := svc.ListFolders(gameID, userID, "note")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Public", result[0].Name)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_RenameFolder_Success(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	folder := models.Folder{ID: folderID, GameID: gameID, UserID: &userID, FolderType: "note"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updated := models.Folder{ID: folderID, Name: "New Name"}

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("Update", folderID, map[string]interface{}{"name": "New Name"}).Return(updated, nil)

	result, err := svc.RenameFolder(folderID, userID, "New Name")

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_RenameFolder_FindByIDError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	folderID := uuid.New()
	userID := uuid.New()

	mockFolderRepo.On("FindByID", folderID).Return(models.Folder{}, gorm.ErrRecordNotFound)

	_, err := svc.RenameFolder(folderID, userID, "New Name")

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockFolderRepo.AssertExpectations(t)
}

func TestFolderService_RenameFolder_PermissionError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	// Session folder, non-GM user
	folder := models.Folder{ID: folderID, GameID: gameID, FolderType: "session"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.RenameFolder(folderID, userID, "New Name")

	assert.ErrorIs(t, err, ErrSessionFoldersReadOnly)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_RenameFolder_EmptyName(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	folder := models.Folder{ID: folderID, GameID: gameID, UserID: &userID, FolderType: "note"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.RenameFolder(folderID, userID, "")

	assert.ErrorIs(t, err, ErrValidation)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_RenameFolder_Update23505Conflict(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	folder := models.Folder{ID: folderID, GameID: gameID, UserID: &userID, FolderType: "note"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	dupErr := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("Update", folderID, map[string]interface{}{"name": "Dup Name"}).Return(models.Folder{}, dupErr)

	_, err := svc.RenameFolder(folderID, userID, "Dup Name")

	assert.ErrorIs(t, err, ErrConflict)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_RenameFolder_UpdateOtherError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	folder := models.Folder{ID: folderID, GameID: gameID, UserID: &userID, FolderType: "note"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	updateErr := errors.New("generic db error")

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("Update", folderID, map[string]interface{}{"name": "New Name"}).Return(models.Folder{}, updateErr)

	_, err := svc.RenameFolder(folderID, userID, "New Name")

	assert.ErrorIs(t, err, updateErr)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_DeleteFolder_Success(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	folder := models.Folder{ID: folderID, GameID: gameID, UserID: &userID, FolderType: "note"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("Delete", folderID).Return(nil)

	err := svc.DeleteFolder(folderID, userID)

	assert.NoError(t, err)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_DeleteFolder_FindByIDError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	folderID := uuid.New()
	userID := uuid.New()

	mockFolderRepo.On("FindByID", folderID).Return(models.Folder{}, gorm.ErrRecordNotFound)

	err := svc.DeleteFolder(folderID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockFolderRepo.AssertExpectations(t)
}

func TestFolderService_DeleteFolder_PermissionError_NoteWrongOwner(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	folderID := uuid.New()
	gameID := uuid.New()
	// Note folder owned by another user
	folder := models.Folder{ID: folderID, GameID: gameID, UserID: &otherUserID, FolderType: "note"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	err := svc.DeleteFolder(folderID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ReorderFolders_Success(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	folderID1 := uuid.New()
	folderID2 := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	scopeFolders := []models.Folder{
		{ID: folderID1, UserID: &userID},
		{ID: folderID2, UserID: &userID},
	}
	orderedIDs := []uuid.UUID{folderID2, folderID1}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindNoteFolders", gameID, userID).Return(scopeFolders, nil)
	mockFolderRepo.On("BatchUpdatePositions", orderedIDs, []int{0, 1}).Return(nil)

	err := svc.ReorderFolders(gameID, userID, "note", orderedIDs)

	assert.NoError(t, err)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ReorderFolders_Forbidden(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.ReorderFolders(gameID, userID, "note", []uuid.UUID{uuid.New()})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ReorderFolders_SessionNonGM(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	err := svc.ReorderFolders(gameID, userID, "session", []uuid.UUID{uuid.New()})

	assert.ErrorIs(t, err, ErrSessionFoldersReadOnly)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ReorderFolders_RepoLoadError(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	repoErr := errors.New("db error")

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindNoteFolders", gameID, userID).Return([]models.Folder{}, repoErr)

	err := svc.ReorderFolders(gameID, userID, "note", []uuid.UUID{uuid.New()})

	assert.ErrorIs(t, err, repoErr)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestFolderService_ReorderFolders_IDNotInScope(t *testing.T) {
	mockFolderRepo := &mocks.MockFolderRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewFolderService(mockFolderRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	folderID := uuid.New()
	outsideID := uuid.New() // Not in the scope
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	scopeFolders := []models.Folder{{ID: folderID, UserID: &userID}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindNoteFolders", gameID, userID).Return(scopeFolders, nil)

	err := svc.ReorderFolders(gameID, userID, "note", []uuid.UUID{outsideID})

	assert.ErrorIs(t, err, ErrForbidden)
	mockFolderRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
