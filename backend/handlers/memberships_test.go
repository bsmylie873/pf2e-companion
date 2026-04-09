package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	authmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

func TestMembershipHandler_CreateMembership_Success(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	userID := uuid.New()

	body := `{"game_id":"` + gameID.String() + `","user_id":"` + userID.String() + `","is_gm":false}`
	req := httptest.NewRequest(http.MethodPost, "/memberships", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.GameMembership{ID: uuid.New(), GameID: gameID, UserID: userID}
	mockSvc.On("CreateMembership", mock.Anything, authUserID).Return(expected, nil)

	err := h.CreateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_CreateMembership_MissingFields(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/memberships", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestMembershipHandler_CreateMembership_InvalidGameID(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	userID := uuid.New()

	body := `{"game_id":"not-a-uuid","user_id":"` + userID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/memberships", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMembershipHandler_ListMemberships_Success(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/memberships?game_id="+gameID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	memberships := []models.GameMembership{{ID: uuid.New()}}
	mockSvc.On("ListMemberships", gameID, authUserID).Return(memberships, nil)

	err := h.ListMemberships(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_ListMemberships_MissingGameID(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/memberships", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ListMemberships(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMembershipHandler_GetMembership_Success(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.GameMembership{ID: id}
	mockSvc.On("GetMembership", id, authUserID).Return(expected, nil)

	err := h.GetMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_GetMembership_NotFound(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetMembership", id, authUserID).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := h.GetMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_UpdateMembership_Success(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"is_gm":true}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.GameMembership{ID: id, IsGM: true}
	mockSvc.On("UpdateMembership", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_DeleteMembership_Success(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteMembership", id, authUserID).Return(nil)

	err := h.DeleteMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_DeleteMembership_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteMembership", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_CreateMembership_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	userID := uuid.New()

	body := `{"game_id":"` + gameID.String() + `","user_id":"` + userID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/memberships", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMembership", mock.Anything, authUserID).Return(models.GameMembership{}, services.ErrForbidden)

	err := h.CreateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_CreateMembership_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	userID := uuid.New()

	body := `{"game_id":"` + gameID.String() + `","user_id":"` + userID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/memberships", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMembership", mock.Anything, authUserID).Return(models.GameMembership{}, errors.New("db error"))

	err := h.CreateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_CreateMembership_InvalidUserID(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"game_id":"` + gameID.String() + `","user_id":"not-a-uuid"}`
	req := httptest.NewRequest(http.MethodPost, "/memberships", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMembershipHandler_ListMemberships_InvalidGameID(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/memberships?game_id=not-a-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ListMemberships(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMembershipHandler_ListMemberships_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/memberships?game_id="+gameID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListMemberships", gameID, authUserID).Return([]models.GameMembership(nil), services.ErrForbidden)

	err := h.ListMemberships(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_ListMemberships_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/memberships?game_id="+gameID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListMemberships", gameID, authUserID).Return([]models.GameMembership(nil), errors.New("db error"))

	err := h.ListMemberships(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_GetMembership_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetMembership", id, authUserID).Return(models.GameMembership{}, services.ErrForbidden)

	err := h.GetMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_GetMembership_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetMembership", id, authUserID).Return(models.GameMembership{}, errors.New("db error"))

	err := h.GetMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_UpdateMembership_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"is_gm":true}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateMembership", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.GameMembership{}, services.ErrForbidden)

	err := h.UpdateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_UpdateMembership_NotFound(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"is_gm":true}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateMembership", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.GameMembership{}, gorm.ErrRecordNotFound)

	err := h.UpdateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_UpdateMembership_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"is_gm":true}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateMembership", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.GameMembership{}, errors.New("db error"))

	err := h.UpdateMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_DeleteMembership_NotFound(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteMembership", id, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DeleteMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMembershipHandler_DeleteMembership_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMembershipService{}
	h := NewMembershipHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteMembership", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteMembership(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
