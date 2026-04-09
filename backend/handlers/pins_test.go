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

func newPinHandler(svc *mocks.MockPinService) *PinHandler {
	return NewPinHandler(svc, NewGameEventHub())
}

func TestPinHandler_CreatePin_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()
	gameID := uuid.New()

	body := `{"label":"Cave","x":0.5,"y":0.5,"colour":"red","icon":"castle"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.SessionPin{ID: uuid.New(), Label: "Cave", GameID: gameID}
	mockSvc.On("CreatePin", sessionID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreatePin_InvalidColour(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	body := `{"label":"Cave","x":0.5,"y":0.5,"colour":"pink","icon":"castle"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPinHandler_CreatePin_InvalidIcon(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	body := `{"label":"Cave","x":0.5,"y":0.5,"colour":"red","icon":"invalid-icon"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPinHandler_CreateGamePin_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"label":"Town","x":0.3,"y":0.7}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.SessionPin{ID: uuid.New(), Label: "Town", GameID: gameID}
	mockSvc.On("CreateGamePin", gameID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateGamePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_ListGamePins_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	pins := []models.SessionPin{{ID: uuid.New()}}
	mockSvc.On("ListGamePins", gameID, authUserID).Return(pins, nil)

	err := h.ListGamePins(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_GetPin_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.SessionPin{ID: id}
	mockSvc.On("GetPin", id, authUserID).Return(expected, nil)

	err := h.GetPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_GetPin_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetPin", id, authUserID).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	err := h.GetPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_UpdatePin_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()
	gameID := uuid.New()

	body := `{"label":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.SessionPin{ID: id, Label: "Updated", GameID: gameID}
	mockSvc.On("UpdatePin", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_UpdatePin_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"label":"X"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdatePin", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.SessionPin{}, services.ErrForbidden)

	err := h.UpdatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_DeletePin_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
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

	pin := models.SessionPin{ID: id, GameID: gameID}
	mockSvc.On("GetPin", id, authUserID).Return(pin, nil)
	mockSvc.On("DeletePin", id, authUserID).Return(nil)

	err := h.DeletePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreateMapPin_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"label":"Camp","x":0.1,"y":0.2}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.SessionPin{ID: uuid.New(), GameID: gameID}
	mockSvc.On("CreateMapPin", mapID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateMapPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_ListMapPins_Success(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	pins := []models.SessionPin{}
	mockSvc.On("ListMapPins", mapID, authUserID).Return(pins, nil)

	err := h.ListMapPins(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreatePin_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	body := `{"label":"Cave","x":0.5,"y":0.5}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreatePin", sessionID, authUserID, mock.Anything).Return(models.SessionPin{}, services.ErrForbidden)

	err := h.CreatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreatePin_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	body := `{"label":"Cave","x":0.5,"y":0.5}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreatePin", sessionID, authUserID, mock.Anything).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	err := h.CreatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreatePin_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	body := `{"label":"Cave","x":0.5,"y":0.5}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreatePin", sessionID, authUserID, mock.Anything).Return(models.SessionPin{}, errors.New("db error"))

	err := h.CreatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreateGamePin_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"label":"Town","x":0.3,"y":0.7}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateGamePin", gameID, authUserID, mock.Anything).Return(models.SessionPin{}, services.ErrForbidden)

	err := h.CreateGamePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreateGamePin_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"label":"Town","x":0.3,"y":0.7}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateGamePin", gameID, authUserID, mock.Anything).Return(models.SessionPin{}, errors.New("db error"))

	err := h.CreateGamePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_ListGamePins_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGamePins", gameID, authUserID).Return([]models.SessionPin(nil), services.ErrForbidden)

	err := h.ListGamePins(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_ListGamePins_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGamePins", gameID, authUserID).Return([]models.SessionPin(nil), errors.New("db error"))

	err := h.ListGamePins(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_GetPin_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetPin", id, authUserID).Return(models.SessionPin{}, services.ErrForbidden)

	err := h.GetPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_GetPin_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetPin", id, authUserID).Return(models.SessionPin{}, errors.New("db error"))

	err := h.GetPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_UpdatePin_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"label":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdatePin", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	err := h.UpdatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_UpdatePin_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"label":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdatePin", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.SessionPin{}, errors.New("db error"))

	err := h.UpdatePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_DeletePin_GetPinForbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetPin", id, authUserID).Return(models.SessionPin{}, services.ErrForbidden)

	err := h.DeletePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_DeletePin_GetPinInternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetPin", id, authUserID).Return(models.SessionPin{}, errors.New("db error"))

	err := h.DeletePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_DeletePin_DeleteForbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
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

	pin := models.SessionPin{ID: id, GameID: gameID}
	mockSvc.On("GetPin", id, authUserID).Return(pin, nil)
	mockSvc.On("DeletePin", id, authUserID).Return(services.ErrForbidden)

	err := h.DeletePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_DeletePin_DeleteInternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
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

	pin := models.SessionPin{ID: id, GameID: gameID}
	mockSvc.On("GetPin", id, authUserID).Return(pin, nil)
	mockSvc.On("DeletePin", id, authUserID).Return(errors.New("db error"))

	err := h.DeletePin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreateMapPin_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"label":"Camp","x":0.1,"y":0.2}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMapPin", mapID, authUserID, mock.Anything).Return(models.SessionPin{}, services.ErrForbidden)

	err := h.CreateMapPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreateMapPin_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"label":"Camp","x":0.1,"y":0.2}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMapPin", mapID, authUserID, mock.Anything).Return(models.SessionPin{}, gorm.ErrRecordNotFound)

	err := h.CreateMapPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_CreateMapPin_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"label":"Camp","x":0.1,"y":0.2}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMapPin", mapID, authUserID, mock.Anything).Return(models.SessionPin{}, errors.New("db error"))

	err := h.CreateMapPin(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_ListMapPins_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListMapPins", mapID, authUserID).Return([]models.SessionPin(nil), services.ErrForbidden)

	err := h.ListMapPins(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinHandler_ListMapPins_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinService{}
	h := newPinHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListMapPins", mapID, authUserID).Return([]models.SessionPin(nil), errors.New("db error"))

	err := h.ListMapPins(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
