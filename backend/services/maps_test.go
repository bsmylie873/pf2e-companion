package services

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestMapService_CreateMap_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("FindActiveByGameID", gameID).Return([]models.GameMap{}, nil)
	mockMapRepo.On("Create", mock.AnythingOfType("*models.GameMap")).Return(nil)

	result, err := svc.CreateMap(gameID, userID, "World Map", nil)

	assert.NoError(t, err)
	assert.Equal(t, "World Map", result.Name)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_CreateMap_NotGM_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateMap(gameID, userID, "World Map", nil)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_CreateMap_DuplicateName_Conflict(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("FindActiveByGameID", gameID).Return([]models.GameMap{}, nil)
	mockMapRepo.On("Create", mock.AnythingOfType("*models.GameMap")).Return(errors.New("23505 duplicate key"))

	_, err := svc.CreateMap(gameID, userID, "World Map", nil)

	assert.ErrorIs(t, err, ErrConflict)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ListMaps_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID}
	maps := []models.GameMap{{ID: uuid.New(), Name: "Map 1"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("FindActiveByGameID", gameID).Return(maps, nil)

	result, err := svc.ListMaps(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_GetMap_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID, Name: "World"}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	result, err := svc.GetMap(mapID, userID)

	assert.NoError(t, err)
	assert.Equal(t, mapID, result.ID)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RenameMap_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID, Name: "Old Name"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	updated := models.GameMap{ID: mapID, Name: "New Name"}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	result, err := svc.RenameMap(mapID, userID, "New Name", nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ArchiveMap_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Archive", mapID).Return(nil)

	err := svc.ArchiveMap(mapID, userID)

	assert.NoError(t, err)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RestoreMap_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	now := time.Now()
	archivedRecently := now.Add(-1 * time.Hour)
	m := models.GameMap{ID: mapID, GameID: gameID, ArchivedAt: &archivedRecently}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	restored := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, map[string]interface{}{"archived_at": nil}).Return(restored, nil)

	result, err := svc.RestoreMap(mapID, userID)

	assert.NoError(t, err)
	assert.Equal(t, mapID, result.ID)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RestoreMap_NotArchived(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID, ArchivedAt: nil}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.RestoreMap(mapID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_CleanupExpiredMaps_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	expiredMapID := uuid.New()
	expiredMaps := []models.GameMap{{ID: expiredMapID, Name: "Old Map"}}

	mockMapRepo.On("FindExpiredArchived", mock.AnythingOfType("time.Time")).Return(expiredMaps, nil)
	mockMapRepo.On("HardDelete", expiredMapID).Return(nil)

	err := svc.CleanupExpiredMaps()

	assert.NoError(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_CleanupExpiredMaps_FindError(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	mockMapRepo.On("FindExpiredArchived", mock.AnythingOfType("time.Time")).Return([]models.GameMap{}, errors.New("db error"))

	err := svc.CleanupExpiredMaps()

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_CleanupExpiredMaps_HardDeleteError(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	expiredMapID := uuid.New()
	expiredMaps := []models.GameMap{{ID: expiredMapID, Name: "Old Map"}}

	mockMapRepo.On("FindExpiredArchived", mock.AnythingOfType("time.Time")).Return(expiredMaps, nil)
	mockMapRepo.On("HardDelete", expiredMapID).Return(errors.New("delete error"))

	err := svc.CleanupExpiredMaps()

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_CleanupExpiredMaps_WithImageURL(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	imageURL := "/uploads/maps/test.jpg"
	expiredMapID := uuid.New()
	expiredMaps := []models.GameMap{{ID: expiredMapID, Name: "Old Map", ImageURL: &imageURL}}

	mockMapRepo.On("FindExpiredArchived", mock.AnythingOfType("time.Time")).Return(expiredMaps, nil)
	mockMapRepo.On("HardDelete", expiredMapID).Return(nil)

	// The file won't exist but os.Remove is silent on that — no error expected
	err := svc.CleanupExpiredMaps()

	assert.NoError(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_ListMaps_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, errors.New("not a member"))

	_, err := svc.ListMaps(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ListArchivedMaps_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	maps := []models.GameMap{{ID: uuid.New(), Name: "Archived Map"}}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("FindArchivedByGameID", gameID).Return(maps, nil)

	result, err := svc.ListArchivedMaps(gameID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ListArchivedMaps_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.ListArchivedMaps(gameID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_GetMap_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, errors.New("not a member"))

	_, err := svc.GetMap(mapID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_GetMap_NotFound(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	_, err := svc.GetMap(mapID, userID)

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_RenameMap_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.RenameMap(mapID, userID, "New Name", nil)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RenameMap_NotFound(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	_, err := svc.RenameMap(mapID, userID, "New Name", nil)

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_RenameMap_NoUpdates(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID, Name: "Existing"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	// Empty name and nil description → no updates → returns existing map unchanged
	result, err := svc.RenameMap(mapID, userID, "", nil)

	assert.NoError(t, err)
	assert.Equal(t, "Existing", result.Name)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RenameMap_Conflict(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID, Name: "Old"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, mock.AnythingOfType("map[string]interface {}")).Return(models.GameMap{}, errors.New("23505 duplicate key"))

	_, err := svc.RenameMap(mapID, userID, "Duplicate", nil)

	assert.ErrorIs(t, err, ErrConflict)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RenameMap_UpdateError(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID, Name: "Old"}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, mock.AnythingOfType("map[string]interface {}")).Return(models.GameMap{}, errors.New("db error"))

	_, err := svc.RenameMap(mapID, userID, "New Name", nil)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrConflict)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ReorderMaps_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", id1, map[string]interface{}{"sort_order": 0}).Return(models.GameMap{}, nil)
	mockMapRepo.On("Update", id2, map[string]interface{}{"sort_order": 1}).Return(models.GameMap{}, nil)

	err := svc.ReorderMaps(gameID, userID, []uuid.UUID{id1, id2})

	assert.NoError(t, err)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ReorderMaps_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	err := svc.ReorderMaps(gameID, userID, []uuid.UUID{uuid.New()})

	assert.ErrorIs(t, err, ErrForbidden)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ReorderMaps_UpdateError(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	id1 := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", id1, map[string]interface{}{"sort_order": 0}).Return(models.GameMap{}, errors.New("db error"))

	err := svc.ReorderMaps(gameID, userID, []uuid.UUID{id1})

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_ArchiveMap_NotFound(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	err := svc.ArchiveMap(mapID, userID)

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_ArchiveMap_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	err := svc.ArchiveMap(mapID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RestoreMap_NotFound(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	_, err := svc.RestoreMap(mapID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_RestoreMap_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	now := time.Now()
	archivedAt := now.Add(-1 * time.Hour)
	m := models.GameMap{ID: mapID, GameID: gameID, ArchivedAt: &archivedAt}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.RestoreMap(mapID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_RestoreMap_Expired(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	// Archived more than 24 hours ago
	expiredAt := time.Now().Add(-25 * time.Hour)
	m := models.GameMap{ID: mapID, GameID: gameID, ArchivedAt: &expiredAt}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.RestoreMap(mapID, userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_SetMapImage_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	updated := models.GameMap{ID: mapID, GameID: gameID, ImageURL: strPtr("/uploads/maps/new.jpg")}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, map[string]interface{}{"image_url": "/uploads/maps/new.jpg"}).Return(updated, nil)

	result, err := svc.SetMapImage(mapID, userID, "/uploads/maps/new.jpg")

	assert.NoError(t, err)
	assert.NotNil(t, result.ImageURL)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_SetMapImage_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.SetMapImage(mapID, userID, "/uploads/maps/new.jpg")

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_SetMapImage_NotFound(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()

	mockMapRepo.On("FindByID", mapID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	_, err := svc.SetMapImage(mapID, userID, "/uploads/maps/new.jpg")

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
}

func TestMapService_SetMapImage_ReplacesExistingImage(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	oldURL := "/uploads/maps/old.jpg"
	m := models.GameMap{ID: mapID, GameID: gameID, ImageURL: &oldURL}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	updated := models.GameMap{ID: mapID, GameID: gameID, ImageURL: strPtr("/uploads/maps/new.jpg")}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, map[string]interface{}{"image_url": "/uploads/maps/new.jpg"}).Return(updated, nil)

	// Old file doesn't exist — os.Remove is silent
	result, err := svc.SetMapImage(mapID, userID, "/uploads/maps/new.jpg")

	assert.NoError(t, err)
	assert.NotNil(t, result.ImageURL)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_DeleteMapImage_Success(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	updated := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, map[string]interface{}{"image_url": nil}).Return(updated, nil)

	result, err := svc.DeleteMapImage(mapID, userID)

	assert.NoError(t, err)
	assert.Nil(t, result.ImageURL)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_DeleteMapImage_Forbidden(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	m := models.GameMap{ID: mapID, GameID: gameID}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: false}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.DeleteMapImage(mapID, userID)

	assert.ErrorIs(t, err, ErrForbidden)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_DeleteMapImage_WithExistingImage(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	mapID := uuid.New()
	gameID := uuid.New()
	existingURL := "/uploads/maps/existing.jpg"
	m := models.GameMap{ID: mapID, GameID: gameID, ImageURL: &existingURL}
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}
	updated := models.GameMap{ID: mapID, GameID: gameID}

	mockMapRepo.On("FindByID", mapID).Return(m, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("Update", mapID, map[string]interface{}{"image_url": nil}).Return(updated, nil)

	// os.Remove is silent when file doesn't exist
	result, err := svc.DeleteMapImage(mapID, userID)

	assert.NoError(t, err)
	assert.Nil(t, result.ImageURL)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_CreateMap_EmptyName(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)

	_, err := svc.CreateMap(gameID, userID, "   ", nil)

	assert.ErrorIs(t, err, ErrValidation)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_CreateMap_FindActiveError(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("FindActiveByGameID", gameID).Return([]models.GameMap{}, errors.New("db error"))

	_, err := svc.CreateMap(gameID, userID, "New Map", nil)

	assert.Error(t, err)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestMapService_CreateMap_CreateError(t *testing.T) {
	mockMapRepo := &mocks.MockMapRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewMapService(mockMapRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	membership := models.GameMembership{UserID: userID, GameID: gameID, IsGM: true}

	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockMapRepo.On("FindActiveByGameID", gameID).Return([]models.GameMap{}, nil)
	mockMapRepo.On("Create", mock.AnythingOfType("*models.GameMap")).Return(errors.New("connection error"))

	_, err := svc.CreateMap(gameID, userID, "New Map", nil)

	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrConflict)
	mockMapRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}
