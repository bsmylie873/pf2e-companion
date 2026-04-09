package handlers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
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

func newMapHandler(svc *mocks.MockMapService) *MapHandler {
	return NewMapHandler(svc, NewGameEventHub())
}

func TestMapHandler_ListMaps_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	maps := []models.GameMap{{ID: uuid.New(), Name: "World Map"}}
	mockSvc.On("ListMaps", gameID, authUserID).Return(maps, nil)

	err := h.ListMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ListArchivedMaps_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	maps := []models.GameMap{}
	mockSvc.On("ListArchivedMaps", gameID, authUserID).Return(maps, nil)

	err := h.ListArchivedMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_CreateMap_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Dungeon"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.GameMap{ID: uuid.New(), Name: "Dungeon", GameID: gameID}
	mockSvc.On("CreateMap", gameID, authUserID, "Dungeon", (*string)(nil)).Return(expected, nil)

	err := h.CreateMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_CreateMap_MissingName(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
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

	err := h.CreateMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestMapHandler_CreateMap_Conflict(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Dungeon"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMap", gameID, authUserID, "Dungeon", (*string)(nil)).Return(models.GameMap{}, services.ErrConflict)

	err := h.CreateMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RenameMap_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"name":"Renamed Map"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.GameMap{ID: mapID, Name: "Renamed Map", GameID: gameID}
	mockSvc.On("RenameMap", mapID, authUserID, "Renamed Map", (*string)(nil)).Return(updated, nil)

	err := h.RenameMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ArchiveMap_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ArchiveMap", mapID, authUserID).Return(nil)

	err := h.ArchiveMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ArchiveMap_NotFound(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ArchiveMap", mapID, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.ArchiveMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RestoreMap_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	restored := models.GameMap{ID: mapID, Name: "Restored", GameID: gameID}
	mockSvc.On("RestoreMap", mapID, authUserID).Return(restored, nil)

	err := h.RestoreMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ReorderMaps_Success(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	m1 := uuid.New()
	m2 := uuid.New()

	body := `{"map_ids":["` + m1.String() + `","` + m2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ReorderMaps", gameID, authUserID, mock.Anything).Return(nil)

	err := h.ReorderMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ReorderMaps_EmptyIDs(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"map_ids":[]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ReorderMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestMapHandler_ListMaps_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListMaps", gameID, authUserID).Return([]models.GameMap{}, services.ErrForbidden)

	err := h.ListMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ListMaps_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListMaps", gameID, authUserID).Return([]models.GameMap{}, errors.New("db error"))

	err := h.ListMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ListArchivedMaps_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListArchivedMaps", gameID, authUserID).Return([]models.GameMap{}, services.ErrForbidden)

	err := h.ListArchivedMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_CreateMap_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMap", gameID, authUserID, "Test", (*string)(nil)).Return(models.GameMap{}, services.ErrForbidden)

	err := h.CreateMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_CreateMap_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateMap", gameID, authUserID, "Test", (*string)(nil)).Return(models.GameMap{}, errors.New("db error"))

	err := h.CreateMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RenameMap_NotFound(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameMap", mapID, authUserID, "New Name", (*string)(nil)).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	err := h.RenameMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RenameMap_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameMap", mapID, authUserID, "New Name", (*string)(nil)).Return(models.GameMap{}, services.ErrForbidden)

	err := h.RenameMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ArchiveMap_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ArchiveMap", mapID, authUserID).Return(services.ErrForbidden)

	err := h.ArchiveMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ArchiveMap_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ArchiveMap", mapID, authUserID).Return(errors.New("db error"))

	err := h.ArchiveMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RestoreMap_NotFound(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RestoreMap", mapID, authUserID).Return(models.GameMap{}, gorm.ErrRecordNotFound)

	err := h.RestoreMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RestoreMap_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RestoreMap", mapID, authUserID).Return(models.GameMap{}, services.ErrForbidden)

	err := h.RestoreMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_RestoreMap_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RestoreMap", mapID, authUserID).Return(models.GameMap{}, errors.New("db error"))

	err := h.RestoreMap(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ReorderMaps_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	m1 := uuid.New()

	body := `{"map_ids":["` + m1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ReorderMaps", gameID, authUserID, mock.Anything).Return(services.ErrForbidden)

	err := h.ReorderMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_ReorderMaps_InternalError(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	m1 := uuid.New()

	body := `{"map_ids":["` + m1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ReorderMaps", gameID, authUserID, mock.Anything).Return(errors.New("db error"))

	err := h.ReorderMaps(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestMapHandler_UploadMapImage_MissingFile(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.UploadMapImage(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMapHandler_UploadMapImage_UnsupportedType(t *testing.T) {
	mockSvc := &mocks.MockMapService{}
	h := newMapHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	mapID := uuid.New()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	mh := make(textproto.MIMEHeader)
	mh.Set("Content-Disposition", `form-data; name="file"; filename="test.txt"`)
	mh.Set("Content-Type", "text/plain")
	part, _ := writer.CreatePart(mh)
	part.Write([]byte("not an image"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "mapId")
	c.SetParamValues(gameID.String(), mapID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.UploadMapImage(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
