package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
)

func TestNoteService_CreateNote_Success(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	note := &models.Note{Title: "My Note"}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("Create", note).Return(nil)

	result, err := svc.CreateNote(gameID, userID, note)

	assert.NoError(t, err)
	assert.Equal(t, "My Note", result.Title)
	assert.Equal(t, gameID, result.GameID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "private", result.Visibility)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_CreateNote_Forbidden(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.CreateNote(gameID, userID, &models.Note{Title: "Note"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_ListGameNotes_Success_GM(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	notes := []models.Note{{ID: uuid.New(), Title: "Note 1"}}
	filters := repositories.NoteFilters{}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("FindByGameID", gameID, userID, true, filters).Return(notes, nil)

	result, err := svc.ListGameNotes(gameID, userID, filters)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_ListGameNotes_Forbidden(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.ListGameNotes(gameID, userID, repositories.NoteFilters{})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_ListGameNotesPaginated_Success(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	notes := []models.Note{{ID: uuid.New(), Title: "N1"}}
	filters := repositories.NoteFilters{}
	var total int64 = 1

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("FindByGameIDPaginated", gameID, userID, false, filters, 0, 10).Return(notes, total, nil)

	result, count, err := svc.ListGameNotesPaginated(gameID, userID, filters, 0, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), count)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_Success_GM(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, Visibility: "private", UserID: uuid.New()}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetNote(noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, noteID, result.ID)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_PrivateNote_NotOwner_NonGM(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	// Private note owned by a different user
	note := models.Note{ID: noteID, GameID: gameID, Visibility: "private", UserID: uuid.New()}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.GetNote(noteID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_OwnPrivateNote(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, Visibility: "private", UserID: userID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetNote(noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, noteID, result.ID)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_Success_Author(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: userID, Visibility: "private"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	updates := map[string]interface{}{"title": "Updated"}
	updatedNote := models.Note{ID: noteID, Title: "Updated"}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("Update", noteID, mock.Anything).Return(updatedNote, nil)

	result, err := svc.UpdateNote(noteID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "Updated", result.Title)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_DeleteNote_Success(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: userID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("ClearNoteFromPins", noteID).Return(nil)
	mockNoteRepo.On("Delete", noteID).Return(nil)

	err := svc.DeleteNote(noteID, userID)

	assert.NoError(t, err)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_DeleteNote_Forbidden_NonOwnerNonGM(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New()}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	err := svc.DeleteNote(noteID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_CreateNote_RepoError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	note := &models.Note{Title: "Fail Note"}
	dbErr := gorm.ErrInvalidDB

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("Create", note).Return(dbErr)

	_, err := svc.CreateNote(gameID, userID, note)

	assert.ErrorIs(t, err, dbErr)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_ListGameNotesPaginated_Forbidden(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, _, err := svc.ListGameNotesPaginated(gameID, userID, repositories.NoteFilters{}, 0, 10)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_RepoError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	noteID := uuid.New()
	userID := uuid.New()

	mockNoteRepo.On("FindByID", noteID).Return(models.Note{}, gorm.ErrRecordNotFound)

	_, err := svc.GetNote(noteID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_MembershipError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.GetNote(noteID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_VisibleNote_NonGM(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, Visibility: "visible", UserID: uuid.New()}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetNote(noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, noteID, result.ID)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_GetNote_EditableNote_NonGM(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, Visibility: "editable", UserID: uuid.New()}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetNote(noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, noteID, result.ID)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_RepoError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	noteID := uuid.New()
	userID := uuid.New()

	mockNoteRepo.On("FindByID", noteID).Return(models.Note{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdateNote(noteID, userID, map[string]interface{}{"title": "x"})

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_MembershipError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New()}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	_, err := svc.UpdateNote(noteID, userID, map[string]interface{}{"title": "x"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_NonAuthor_NonGM_Editable(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New(), Visibility: "editable"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	// Include "visibility" in updates — should be stripped because non-author non-GM
	updates := map[string]interface{}{"title": "updated", "visibility": "private"}
	updatedNote := models.Note{ID: noteID, Title: "updated"}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("Update", noteID, mock.Anything).Return(updatedNote, nil)

	result, err := svc.UpdateNote(noteID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "updated", result.Title)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_NonAuthor_NonGM_Private_Forbidden(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New(), Visibility: "private"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.UpdateNote(noteID, userID, map[string]interface{}{"title": "x"})

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_GM_NonAuthor_StripsVisibility(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New(), Visibility: "private"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	// Include "visibility" in updates — should be stripped because GM non-author
	updates := map[string]interface{}{"title": "gm edit", "visibility": "visible"}
	updatedNote := models.Note{ID: noteID, Title: "gm edit"}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("Update", noteID, mock.Anything).Return(updatedNote, nil)

	result, err := svc.UpdateNote(noteID, userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "gm edit", result.Title)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_FolderVisibilityValidation_PrivateFolder(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	folderID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: userID, Visibility: "private", FolderID: &folderID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	folder := models.Folder{ID: folderID, Visibility: "private"}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindByID", folderID).Return(folder, nil)

	_, err := svc.UpdateNote(noteID, userID, map[string]interface{}{"visibility": "editable"})

	assert.ErrorIs(t, err, ErrValidation)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
}

func TestNoteService_UpdateNote_FolderIDChange_PrivateFolderNonPrivateNote(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	newFolderID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: userID, Visibility: "visible"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	folder := models.Folder{ID: newFolderID, Visibility: "private"}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockFolderRepo.On("FindByID", newFolderID).Return(folder, nil)

	_, err := svc.UpdateNote(noteID, userID, map[string]interface{}{"folder_id": newFolderID})

	assert.ErrorIs(t, err, ErrValidation)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
}

func TestNoteService_DeleteNote_RepoError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	noteID := uuid.New()
	userID := uuid.New()

	mockNoteRepo.On("FindByID", noteID).Return(models.Note{}, gorm.ErrRecordNotFound)

	err := svc.DeleteNote(noteID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockNoteRepo.AssertExpectations(t)
}

func TestNoteService_DeleteNote_MembershipError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New()}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := svc.DeleteNote(noteID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_DeleteNote_GM_NonOwner(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: uuid.New()}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("ClearNoteFromPins", noteID).Return(nil)
	mockNoteRepo.On("Delete", noteID).Return(nil)

	err := svc.DeleteNote(noteID, userID)

	assert.NoError(t, err)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestNoteService_DeleteNote_ClearNoteFromPinsError(t *testing.T) {
	mockNoteRepo := &mocks.MockNoteRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	mockFolderRepo := &mocks.MockFolderRepository{}
	svc := NewNoteService(mockNoteRepo, mockMemberRepo, mockFolderRepo)

	userID := uuid.New()
	noteID := uuid.New()
	gameID := uuid.New()
	note := models.Note{ID: noteID, GameID: gameID, UserID: userID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}
	clearErr := gorm.ErrInvalidDB

	mockNoteRepo.On("FindByID", noteID).Return(note, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockNoteRepo.On("ClearNoteFromPins", noteID).Return(clearErr)

	err := svc.DeleteNote(noteID, userID)

	assert.ErrorIs(t, err, clearErr)
	mockNoteRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
