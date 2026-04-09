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

func newSessionHandler(svc *mocks.MockSessionService) *SessionHandler {
	return NewSessionHandler(svc, NewGameEventHub())
}

func TestSessionHandler_CreateSession_Success(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"Session 1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Session{ID: uuid.New(), Title: "Session 1", GameID: gameID}
	mockSvc.On("CreateSession", gameID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_CreateSession_MissingTitle(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSessionHandler_CreateSession_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"Session 1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateSession", gameID, authUserID, mock.Anything).Return(models.Session{}, services.ErrForbidden)

	err := h.CreateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_ListGameSessions_Success(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	sessions := []models.Session{{ID: uuid.New(), Title: "Session 1"}}
	mockSvc.On("ListGameSessions", gameID, authUserID).Return(sessions, nil)

	err := h.ListGameSessions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_GetSession_Success(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Session{ID: id, Title: "My Session"}
	mockSvc.On("GetSession", id, authUserID).Return(expected, nil)

	err := h.GetSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_GetSession_NotFound(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetSession", id, authUserID).Return(models.Session{}, gorm.ErrRecordNotFound)

	err := h.GetSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_UpdateSession_Success(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()
	gameID := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.Session{ID: id, Title: "Updated", GameID: gameID}
	mockSvc.On("UpdateSession", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_Success(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	session := models.Session{ID: id, GameID: gameID}
	mockSvc.On("GetSession", id, authUserID).Return(session, nil)
	mockSvc.On("DeleteSession", id, authUserID).Return(nil)

	err := h.DeleteSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_NotFound(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetSession", id, authUserID).Return(models.Session{}, gorm.ErrRecordNotFound)

	err := h.DeleteSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_CreateSession_InvalidRuntime(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"Session","runtime_start":"2024-01-02T00:00:00Z","runtime_end":"2024-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSessionHandler_CreateSession_InternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"Session 1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateSession", gameID, authUserID, mock.Anything).Return(models.Session{}, errors.New("db error"))

	err := h.CreateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_ListGameSessions_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameSessions", gameID, authUserID).Return([]models.Session(nil), services.ErrForbidden)

	err := h.ListGameSessions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_ListGameSessions_InternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameSessions", gameID, authUserID).Return([]models.Session(nil), errors.New("db error"))

	err := h.ListGameSessions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_GetSession_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetSession", id, authUserID).Return(models.Session{}, services.ErrForbidden)

	err := h.GetSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_GetSession_InternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetSession", id, authUserID).Return(models.Session{}, errors.New("db error"))

	err := h.GetSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_UpdateSession_InvalidRuntime(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"runtime_start":"2024-01-02T00:00:00Z","runtime_end":"2024-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.UpdateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestSessionHandler_UpdateSession_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateSession", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Session{}, services.ErrForbidden)

	err := h.UpdateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_UpdateSession_NotFound(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateSession", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Session{}, gorm.ErrRecordNotFound)

	err := h.UpdateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_UpdateSession_InternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"title":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateSession", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Session{}, errors.New("db error"))

	err := h.UpdateSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_GetSessionForbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetSession", id, authUserID).Return(models.Session{}, services.ErrForbidden)

	err := h.DeleteSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_GetSessionInternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetSession", id, authUserID).Return(models.Session{}, errors.New("db error"))

	err := h.DeleteSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_DeleteForbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	session := models.Session{ID: id, GameID: gameID}
	mockSvc.On("GetSession", id, authUserID).Return(session, nil)
	mockSvc.On("DeleteSession", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_DeleteSession_DeleteInternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := newSessionHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	session := models.Session{ID: id, GameID: gameID}
	mockSvc.On("GetSession", id, authUserID).Return(session, nil)
	mockSvc.On("DeleteSession", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_ListGameSessions_Paginated_Success(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := NewSessionHandler(mockSvc, NewGameEventHub())
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	sessions := []models.Session{}
	mockSvc.On("ListGameSessionsPaginated", gameID, authUserID, 0, 10).Return(sessions, int64(0), nil)

	err := h.ListGameSessions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_ListGameSessions_Paginated_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := NewSessionHandler(mockSvc, NewGameEventHub())
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameSessionsPaginated", gameID, authUserID, 0, 10).Return([]models.Session(nil), int64(0), services.ErrForbidden)

	err := h.ListGameSessions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestSessionHandler_ListGameSessions_Paginated_InternalError(t *testing.T) {
	mockSvc := &mocks.MockSessionService{}
	h := NewSessionHandler(mockSvc, NewGameEventHub())
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameSessionsPaginated", gameID, authUserID, 0, 10).Return([]models.Session(nil), int64(0), errors.New("db error"))

	err := h.ListGameSessions(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
