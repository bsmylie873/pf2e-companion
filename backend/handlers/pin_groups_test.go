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

func newPinGroupHandler(svc *mocks.MockPinGroupService) *PinGroupHandler {
	return NewPinGroupHandler(svc, NewGameEventHub())
}

func TestPinGroupHandler_CreateGroup_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.PinGroupResponse{ID: uuid.New(), GameID: gameID}
	mockSvc.On("CreateGroup", gameID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateGroup_TooFewPins(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	p1 := uuid.New()

	body := `{"pin_ids":["` + p1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPinGroupHandler_ListGameGroups_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	groups := []models.PinGroupResponse{}
	mockSvc.On("ListGameGroups", gameID, authUserID).Return(groups, nil)

	err := h.ListGameGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_UpdateGroup_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()

	body := `{"colour":"blue"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("UpdateGroup", groupID, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_UpdateGroup_InvalidColour(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	body := `{"colour":"magenta"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.UpdateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPinGroupHandler_AddPinToGroup_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()

	body := `{"pin_id":"` + pinID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("AddPinToGroup", groupID, pinID, authUserID).Return(resp, nil)

	err := h.AddPinToGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_RemovePinFromGroup_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "pinId")
	c.SetParamValues(groupID.String(), pinID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("RemovePinFromGroup", groupID, pinID, authUserID).Return(resp, nil)

	err := h.RemovePinFromGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	group := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("GetGroup", groupID, authUserID).Return(group, nil)
	mockSvc.On("DisbandGroup", groupID, authUserID).Return(nil)

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetGroup", groupID, authUserID).Return(models.PinGroupResponse{}, gorm.ErrRecordNotFound)

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateMapGroup_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.PinGroupResponse{ID: uuid.New(), GameID: gameID}
	mockSvc.On("CreateMapGroup", mapID, authUserID, mock.Anything).Return(resp, nil)

	err := h.CreateMapGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_ListMapGroups_Success(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
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

	groups := []models.PinGroupResponse{}
	mockSvc.On("ListMapGroups", mapID, authUserID).Return(groups, nil)

	err := h.ListMapGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateGroup_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateGroup", gameID, authUserID, mock.Anything).Return(models.PinGroupResponse{}, services.ErrForbidden)

	err := h.CreateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateGroup_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateGroup", gameID, authUserID, mock.Anything).Return(models.PinGroupResponse{}, gorm.ErrRecordNotFound)

	err := h.CreateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateGroup_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateGroup", gameID, authUserID, mock.Anything).Return(models.PinGroupResponse{}, errors.New("db error"))

	err := h.CreateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_ListGameGroups_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameGroups", gameID, authUserID).Return([]models.PinGroupResponse(nil), services.ErrForbidden)

	err := h.ListGameGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_ListGameGroups_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameGroups", gameID, authUserID).Return([]models.PinGroupResponse(nil), errors.New("db error"))

	err := h.ListGameGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_UpdateGroup_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	body := `{"colour":"blue"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateGroup", groupID, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.PinGroupResponse{}, services.ErrForbidden)

	err := h.UpdateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_UpdateGroup_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	body := `{"colour":"blue"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateGroup", groupID, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.PinGroupResponse{}, gorm.ErrRecordNotFound)

	err := h.UpdateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_UpdateGroup_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	body := `{"colour":"blue"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateGroup", groupID, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.PinGroupResponse{}, errors.New("db error"))

	err := h.UpdateGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_AddPinToGroup_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	body := `{"pin_id":"` + pinID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("AddPinToGroup", groupID, pinID, authUserID).Return(models.PinGroupResponse{}, services.ErrForbidden)

	err := h.AddPinToGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_AddPinToGroup_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	body := `{"pin_id":"` + pinID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("AddPinToGroup", groupID, pinID, authUserID).Return(models.PinGroupResponse{}, gorm.ErrRecordNotFound)

	err := h.AddPinToGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_AddPinToGroup_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	body := `{"pin_id":"` + pinID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("AddPinToGroup", groupID, pinID, authUserID).Return(models.PinGroupResponse{}, errors.New("db error"))

	err := h.AddPinToGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_RemovePinFromGroup_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "pinId")
	c.SetParamValues(groupID.String(), pinID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RemovePinFromGroup", groupID, pinID, authUserID).Return(models.PinGroupResponse{}, services.ErrForbidden)

	err := h.RemovePinFromGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_RemovePinFromGroup_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "pinId")
	c.SetParamValues(groupID.String(), pinID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RemovePinFromGroup", groupID, pinID, authUserID).Return(models.PinGroupResponse{}, gorm.ErrRecordNotFound)

	err := h.RemovePinFromGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_RemovePinFromGroup_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	pinID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "pinId")
	c.SetParamValues(groupID.String(), pinID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RemovePinFromGroup", groupID, pinID, authUserID).Return(models.PinGroupResponse{}, errors.New("db error"))

	err := h.RemovePinFromGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_GetGroupForbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetGroup", groupID, authUserID).Return(models.PinGroupResponse{}, services.ErrForbidden)

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_GetGroupInternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetGroup", groupID, authUserID).Return(models.PinGroupResponse{}, errors.New("db error"))

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_DisbandForbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	group := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("GetGroup", groupID, authUserID).Return(group, nil)
	mockSvc.On("DisbandGroup", groupID, authUserID).Return(services.ErrForbidden)

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_DisbandNotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	group := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("GetGroup", groupID, authUserID).Return(group, nil)
	mockSvc.On("DisbandGroup", groupID, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_DisbandGroup_DisbandInternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	groupID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(groupID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	group := models.PinGroupResponse{ID: groupID, GameID: gameID}
	mockSvc.On("GetGroup", groupID, authUserID).Return(group, nil)
	mockSvc.On("DisbandGroup", groupID, authUserID).Return(errors.New("db error"))

	err := h.DisbandGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateMapGroup_TooFewPins(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	p1 := uuid.New()

	body := `{"pin_ids":["` + p1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateMapGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestPinGroupHandler_CreateMapGroup_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMapGroup", mapID, authUserID, mock.Anything).Return(models.PinGroupResponse{}, services.ErrForbidden)

	err := h.CreateMapGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateMapGroup_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMapGroup", mapID, authUserID, mock.Anything).Return(models.PinGroupResponse{}, gorm.ErrRecordNotFound)

	err := h.CreateMapGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_CreateMapGroup_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()
	p1, p2 := uuid.New(), uuid.New()

	body := `{"pin_ids":["` + p1.String() + `","` + p2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMapGroup", mapID, authUserID, mock.Anything).Return(models.PinGroupResponse{}, errors.New("db error"))

	err := h.CreateMapGroup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_ListMapGroups_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
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

	mockSvc.On("ListMapGroups", mapID, authUserID).Return([]models.PinGroupResponse(nil), services.ErrForbidden)

	err := h.ListMapGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_ListMapGroups_NotFound(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
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

	mockSvc.On("ListMapGroups", mapID, authUserID).Return([]models.PinGroupResponse(nil), gorm.ErrRecordNotFound)

	err := h.ListMapGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPinGroupHandler_ListMapGroups_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPinGroupService{}
	h := newPinGroupHandler(mockSvc)
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

	mockSvc.On("ListMapGroups", mapID, authUserID).Return([]models.PinGroupResponse(nil), errors.New("db error"))

	err := h.ListMapGroups(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
