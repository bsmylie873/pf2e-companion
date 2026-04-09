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

func TestBackupHandler_ExportGame_Success(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	backup := &models.BackupFile{SchemaVersion: "1", GameID: gameID}
	mockSvc.On("ExportGame", gameID, authUserID).Return(backup, nil)

	err := h.ExportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "attachment")
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportGame_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportGame", gameID, authUserID).Return((*models.BackupFile)(nil), services.ErrForbidden)

	err := h.ExportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportGame_NotFound(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportGame", gameID, authUserID).Return((*models.BackupFile)(nil), gorm.ErrRecordNotFound)

	err := h.ExportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportSession_Success(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	backup := &models.BackupFile{SchemaVersion: "1"}
	mockSvc.On("ExportSession", sessionID, authUserID).Return(backup, nil)

	err := h.ExportSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportNote_Success(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(noteID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	backup := &models.BackupFile{SchemaVersion: "1"}
	mockSvc.On("ExportNote", noteID, authUserID).Return(backup, nil)

	err := h.ExportNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ImportGame_Success(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	backup := models.BackupFile{SchemaVersion: "1", GameID: gameID}
	bodyBytes, _ := json.Marshal(backup)

	req := httptest.NewRequest(http.MethodPost, "/?mode=merge", strings.NewReader(string(bodyBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	summary := &models.ImportSummary{SessionsCreated: 2, NotesCreated: 5}
	mockSvc.On("ImportGame", gameID, authUserID, "merge", mock.Anything).Return(summary, nil)

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ImportGame_InvalidMode(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/?mode=invalid", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBackupHandler_ImportGame_UnsupportedSchemaVersion(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	backup := models.BackupFile{SchemaVersion: "99", GameID: gameID}
	bodyBytes, _ := json.Marshal(backup)

	req := httptest.NewRequest(http.MethodPost, "/?mode=merge", strings.NewReader(string(bodyBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBackupHandler_ImportGame_GameIDMismatch(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	otherGameID := uuid.New()

	backup := models.BackupFile{SchemaVersion: "1", GameID: otherGameID}
	bodyBytes, _ := json.Marshal(backup)

	req := httptest.NewRequest(http.MethodPost, "/?mode=merge", strings.NewReader(string(bodyBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBackupHandler_ExportGame_InternalError(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportGame", gameID, authUserID).Return((*models.BackupFile)(nil), errors.New("db error"))

	err := h.ExportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportSession_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportSession", sessionID, authUserID).Return((*models.BackupFile)(nil), services.ErrForbidden)

	err := h.ExportSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportSession_NotFound(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportSession", sessionID, authUserID).Return((*models.BackupFile)(nil), gorm.ErrRecordNotFound)

	err := h.ExportSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportSession_InternalError(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	sessionID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(sessionID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportSession", sessionID, authUserID).Return((*models.BackupFile)(nil), errors.New("db error"))

	err := h.ExportSession(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportNote_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(noteID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportNote", noteID, authUserID).Return((*models.BackupFile)(nil), services.ErrForbidden)

	err := h.ExportNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportNote_NotFound(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(noteID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportNote", noteID, authUserID).Return((*models.BackupFile)(nil), gorm.ErrRecordNotFound)

	err := h.ExportNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ExportNote_InternalError(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(noteID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ExportNote", noteID, authUserID).Return((*models.BackupFile)(nil), errors.New("db error"))

	err := h.ExportNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ImportGame_InvalidJSON(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/?mode=merge", strings.NewReader("not-json!"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestBackupHandler_ImportGame_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	backup := models.BackupFile{SchemaVersion: "1", GameID: gameID}
	bodyBytes, _ := json.Marshal(backup)

	req := httptest.NewRequest(http.MethodPost, "/?mode=merge", strings.NewReader(string(bodyBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ImportGame", gameID, authUserID, "merge", mock.Anything).Return((*models.ImportSummary)(nil), services.ErrForbidden)

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestBackupHandler_ImportGame_InternalError(t *testing.T) {
	mockSvc := &mocks.MockBackupService{}
	h := NewBackupHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	backup := models.BackupFile{SchemaVersion: "1", GameID: gameID}
	bodyBytes, _ := json.Marshal(backup)

	req := httptest.NewRequest(http.MethodPost, "/?mode=overwrite", strings.NewReader(string(bodyBytes)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ImportGame", gameID, authUserID, "overwrite", mock.Anything).Return((*models.ImportSummary)(nil), errors.New("db error"))

	err := h.ImportGame(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
