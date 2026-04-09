package handlers

import (
	"encoding/json"
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

func setupGameHandler(t *testing.T) (*GameHandler, *mocks.MockGameService, uuid.UUID) {
	t.Helper()
	mockSvc := &mocks.MockGameService{}
	h := NewGameHandler(mockSvc)
	authUserID := uuid.New()
	return h, mockSvc, authUserID
}

func TestGameHandler_CreateGame_Success(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()

	body := `{"title":"Test Game"}`
	req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Game{ID: uuid.New(), Title: "Test Game"}
	mockSvc.On("CreateGame", mock.Anything, mock.Anything, authUserID).Return(expected, nil)

	err := h.CreateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_CreateGame_MissingTitle(t *testing.T) {
	h, _, authUserID := setupGameHandler(t)
	e := echo.New()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestGameHandler_CreateGame_InvalidMemberUUID(t *testing.T) {
	h, _, authUserID := setupGameHandler(t)
	e := echo.New()

	body := `{"title":"Test","members":[{"user_id":"not-a-uuid"}]}`
	req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGameHandler_CreateGame_NoAuth(t *testing.T) {
	h, _, _ := setupGameHandler(t)
	e := echo.New()

	body := `{"title":"Test Game"}`
	req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGameHandler_ListGames_Success(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/games", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	games := []models.Game{{ID: uuid.New(), Title: "Game 1"}}
	mockSvc.On("ListGames", authUserID).Return(games, nil)

	err := h.ListGames(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_ListGames_Paginated(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/games?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	games := []models.Game{{ID: uuid.New(), Title: "Game 1"}}
	mockSvc.On("ListGamesPaginated", authUserID, 0, 10).Return(games, int64(1), nil)

	err := h.ListGames(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_GetGame_Success(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Game{ID: id, Title: "My Game"}
	mockSvc.On("GetGame", id, authUserID).Return(expected, nil)

	err := h.GetGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_GetGame_Forbidden(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetGame", id, authUserID).Return(models.Game{}, services.ErrForbidden)

	err := h.GetGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_GetGame_NotFound(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetGame", id, authUserID).Return(models.Game{}, gorm.ErrRecordNotFound)

	err := h.GetGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_UpdateGame_Success(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/games/"+id.String(), strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.Game{ID: id, Title: "Updated"}
	mockSvc.On("UpdateGame", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_UpdateGame_Forbidden(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/games/"+id.String(), strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateGame", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Game{}, services.ErrForbidden)

	err := h.UpdateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_DeleteGame_Success(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteGame", id, authUserID).Return(nil)

	err := h.DeleteGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_DeleteGame_NotFound(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteGame", id, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DeleteGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_CreateGame_ServiceError(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()

	body := `{"title":"Test Game"}`
	req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateGame", mock.Anything, mock.Anything, authUserID).Return(models.Game{}, errors.New("db error"))

	err := h.CreateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_ListGames_InternalError(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/games", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGames", authUserID).Return([]models.Game(nil), errors.New("db error"))

	err := h.ListGames(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_GetGame_InternalError(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetGame", id, authUserID).Return(models.Game{}, errors.New("db error"))

	err := h.GetGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_UpdateGame_NotFound(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/games/"+id.String(), strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateGame", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Game{}, gorm.ErrRecordNotFound)

	err := h.UpdateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_UpdateGame_InternalError(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/games/"+id.String(), strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateGame", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Game{}, errors.New("db error"))

	err := h.UpdateGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_DeleteGame_Forbidden(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteGame", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestGameHandler_DeleteGame_InternalError(t *testing.T) {
	h, mockSvc, authUserID := setupGameHandler(t)
	e := echo.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/games/"+id.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteGame", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
