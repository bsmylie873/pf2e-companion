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

func TestFolderHandler_CreateFolder_Success(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Chapter 1","folder_type":"session","visibility":"game-wide"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Folder{ID: uuid.New(), Name: "Chapter 1"}
	mockSvc.On("CreateFolder", gameID, authUserID, "Chapter 1", "session", "game-wide").Return(expected, nil)

	err := h.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_CreateFolder_MissingName(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"folder_type":"session"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestFolderHandler_ListFolders_Success(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?type=session", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	folders := []models.Folder{{ID: uuid.New(), Name: "Chapter 1"}}
	mockSvc.On("ListFolders", gameID, authUserID, "session").Return(folders, nil)

	err := h.ListFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ListFolders_InvalidType(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?type=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ListFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFolderHandler_RenameFolder_Success(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Renamed"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.Folder{ID: id, Name: "Renamed"}
	mockSvc.On("RenameFolder", id, authUserID, "Renamed").Return(updated, nil)

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_DeleteFolder_Success(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteFolder", id, authUserID).Return(nil)

	err := h.DeleteFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_DeleteFolder_NotFound(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteFolder", id, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DeleteFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ReorderFolders_Success(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	f1 := uuid.New()
	f2 := uuid.New()

	body := `{"folder_type":"session","folder_ids":["` + f1.String() + `","` + f2.String() + `"]}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ReorderFolders", gameID, authUserID, "session", mock.Anything).Return(nil)

	err := h.ReorderFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ReorderFolders_EmptyIDs(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"folder_type":"session","folder_ids":[]}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ReorderFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFolderHandler_CreateFolder_Conflict(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Dup","folder_type":"session","visibility":"game-wide"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateFolder", gameID, authUserID, "Dup", "session", "game-wide").Return(models.Folder{}, services.ErrConflict)

	err := h.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_CreateFolder_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"New","folder_type":"session","visibility":"game-wide"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateFolder", gameID, authUserID, "New", "session", "game-wide").Return(models.Folder{}, services.ErrForbidden)

	err := h.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_CreateFolder_SessionReadOnly(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"New","folder_type":"session"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateFolder", gameID, authUserID, "New", "session", "game-wide").Return(models.Folder{}, services.ErrSessionFoldersReadOnly)

	err := h.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_CreateFolder_InternalError(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"New","folder_type":"session","visibility":"game-wide"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateFolder", gameID, authUserID, "New", "session", "game-wide").Return(models.Folder{}, errors.New("db error"))

	err := h.CreateFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ListFolders_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?type=note", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListFolders", gameID, authUserID, "note").Return([]models.Folder(nil), services.ErrForbidden)

	err := h.ListFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ListFolders_InternalError(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?type=note", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListFolders", gameID, authUserID, "note").Return([]models.Folder(nil), errors.New("db error"))

	err := h.ListFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_RenameFolder_MissingName(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestFolderHandler_RenameFolder_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameFolder", id, authUserID, "New Name").Return(models.Folder{}, services.ErrForbidden)

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_RenameFolder_NotFound(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameFolder", id, authUserID, "New Name").Return(models.Folder{}, gorm.ErrRecordNotFound)

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_RenameFolder_Conflict(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Dup Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameFolder", id, authUserID, "Dup Name").Return(models.Folder{}, services.ErrConflict)

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_RenameFolder_InternalError(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameFolder", id, authUserID, "New Name").Return(models.Folder{}, errors.New("db error"))

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_DeleteFolder_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteFolder", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_DeleteFolder_InternalError(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteFolder", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ReorderFolders_InvalidType(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	f1 := uuid.New()

	body := `{"folder_type":"invalid","folder_ids":["` + f1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ReorderFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFolderHandler_ReorderFolders_InvalidUUID(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"folder_type":"session","folder_ids":["not-a-uuid"]}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ReorderFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFolderHandler_ReorderFolders_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	f1 := uuid.New()

	body := `{"folder_type":"session","folder_ids":["` + f1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ReorderFolders", gameID, authUserID, "session", mock.Anything).Return(services.ErrForbidden)

	err := h.ReorderFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_ReorderFolders_InternalError(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	f1 := uuid.New()

	body := `{"folder_type":"session","folder_ids":["` + f1.String() + `"]}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ReorderFolders", gameID, authUserID, "session", mock.Anything).Return(errors.New("db error"))

	err := h.ReorderFolders(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestFolderHandler_mapFolderError_Validation(t *testing.T) {
	mockSvc := &mocks.MockFolderService{}
	h := NewFolderHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Bad Vis"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RenameFolder", id, authUserID, "Bad Vis").Return(models.Folder{}, services.ErrValidation)

	err := h.RenameFolder(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}
