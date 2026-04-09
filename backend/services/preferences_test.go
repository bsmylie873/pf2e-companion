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

func TestPreferenceService_GetPreferences_Success(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	pref := models.UserPreference{
		UserID:        userID,
		MapEditorMode: "modal",
	}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	result, err := svc.GetPreferences(userID)

	assert.NoError(t, err)
	assert.Equal(t, "modal", result.MapEditorMode)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_GetPreferences_NotFound_ReturnsEmpty(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()

	mockPrefRepo.On("FindByUserID", userID).Return(models.UserPreference{}, gorm.ErrRecordNotFound)

	result, err := svc.GetPreferences(userID)

	assert.NoError(t, err)
	assert.Equal(t, models.UserPreferenceResponse{}, result)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_GetPreferences_DBError(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	dbErr := errors.New("db error")

	mockPrefRepo.On("FindByUserID", userID).Return(models.UserPreference{}, dbErr)

	_, err := svc.GetPreferences(userID)

	assert.ErrorIs(t, err, dbErr)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_MapEditorMode(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"map_editor_mode": "navigate"}
	pref := models.UserPreference{UserID: userID, MapEditorMode: "modal"}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", &models.UserPreference{UserID: userID, MapEditorMode: "navigate"}).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "navigate", result.MapEditorMode)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_InvalidMapEditorMode(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"map_editor_mode": "invalid"}
	pref := models.UserPreference{UserID: userID, MapEditorMode: "modal"}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	_, err := svc.UpdatePreferences(userID, updates)

	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultPinColour(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_pin_colour": "red"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", &models.UserPreference{UserID: userID, DefaultPinColour: strPtr("red")}).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)

	assert.NoError(t, err)
	assert.NotNil(t, result.DefaultPinColour)
	assert.Equal(t, "red", *result.DefaultPinColour)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_InvalidPinColour(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_pin_colour": "hot-pink"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	_, err := svc.UpdatePreferences(userID, updates)

	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_NewUser_CreatesPreference(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"map_editor_mode": "modal"}

	mockPrefRepo.On("FindByUserID", userID).Return(models.UserPreference{}, gorm.ErrRecordNotFound)
	mockPrefRepo.On("Upsert", &models.UserPreference{UserID: userID, MapEditorMode: "modal"}).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)

	assert.NoError(t, err)
	assert.Equal(t, "modal", result.MapEditorMode)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultGameID_Success(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	updates := map[string]interface{}{"default_game_id": gameID}
	pref := models.UserPreference{UserID: userID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPrefRepo.On("Upsert", &models.UserPreference{UserID: userID, DefaultGameID: &gameID}).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)

	assert.NoError(t, err)
	assert.NotNil(t, result.DefaultGameID)
	assert.Equal(t, gameID, *result.DefaultGameID)
	mockPrefRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func strPtr(s string) *string {
	return &s
}

func TestPreferenceService_UpdatePreferences_DefaultGameID_NilClears(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	updates := map[string]interface{}{"default_game_id": nil}
	existingPref := models.UserPreference{UserID: userID, DefaultGameID: &gameID}

	mockPrefRepo.On("FindByUserID", userID).Return(existingPref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	assert.Nil(t, result.DefaultGameID)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultGameID_StringUUID(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	updates := map[string]interface{}{"default_game_id": gameID.String()}
	pref := models.UserPreference{UserID: userID}
	membership := models.GameMembership{UserID: userID, GameID: gameID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(membership, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	assert.NotNil(t, result.DefaultGameID)
	assert.Equal(t, gameID, *result.DefaultGameID)
	mockPrefRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultGameID_InvalidStringUUID(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_game_id": "not-a-valid-uuid"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultGameID_InvalidType(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_game_id": 42}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultGameID_MembershipError(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()
	updates := map[string]interface{}{"default_game_id": gameID}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockMemberRepo.On("FindByUserAndGameID", userID, gameID).Return(models.GameMembership{}, errors.New("not a member"))

	_, err := svc.UpdatePreferences(userID, updates)
	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultPinColour_NilClears(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_pin_colour": nil}
	colour := "red"
	pref := models.UserPreference{UserID: userID, DefaultPinColour: &colour}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	assert.Nil(t, result.DefaultPinColour)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultPinIcon_ValidIcon(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_pin_icon": "castle"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	result, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	assert.NotNil(t, result.DefaultPinIcon)
	assert.Equal(t, "castle", *result.DefaultPinIcon)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultPinIcon_NilClears(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_pin_icon": nil}
	icon := "castle"
	pref := models.UserPreference{UserID: userID, DefaultPinIcon: &icon}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultPinIcon_InvalidIcon(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_pin_icon": "unicorn-riding-dragon"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_SidebarState_Set(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"sidebar_state": map[string]interface{}{"open": true}}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_SidebarState_NilClears(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"sidebar_state": nil}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultViewMode_Set(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_view_mode": "list"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_DefaultViewMode_NilClears(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"default_view_mode": nil}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_PageSize_Set(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"page_size": map[string]interface{}{"notes": 25}}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_PageSize_NilClears(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"page_size": nil}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_MapEditorMode_NilValue(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"map_editor_mode": nil}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)

	_, err := svc.UpdatePreferences(userID, updates)
	assert.ErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_FindDBError(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"map_editor_mode": "modal"}

	mockPrefRepo.On("FindByUserID", userID).Return(models.UserPreference{}, errors.New("connection refused"))

	_, err := svc.UpdatePreferences(userID, updates)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrValidation)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_UpdatePreferences_UpsertError(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	updates := map[string]interface{}{"map_editor_mode": "modal"}
	pref := models.UserPreference{UserID: userID}

	mockPrefRepo.On("FindByUserID", userID).Return(pref, nil)
	mockPrefRepo.On("Upsert", mock.AnythingOfType("*models.UserPreference")).Return(errors.New("db error"))

	_, err := svc.UpdatePreferences(userID, updates)
	assert.Error(t, err)
	mockPrefRepo.AssertExpectations(t)
}

func TestPreferenceService_ClearDefaultGameForMembership_Success(t *testing.T) {
	mockPrefRepo := &mocks.MockPreferenceRepository{}
	mockMemberRepo := &mocks.MockMembershipRepository{}
	svc := NewPreferenceService(mockPrefRepo, mockMemberRepo)

	userID := uuid.New()
	gameID := uuid.New()

	mockPrefRepo.On("ClearDefaultGameForMembership", userID, gameID).Return(nil)

	err := svc.ClearDefaultGameForMembership(userID, gameID)
	assert.NoError(t, err)
	mockPrefRepo.AssertExpectations(t)
}
