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
	"pf2e-companion/backend/repositories"
	"pf2e-companion/backend/services"
)

func newNoteHandler(svc *mocks.MockNoteService) *NoteHandler {
	return NewNoteHandler(svc, NewGameEventHub())
}

func TestNoteHandler_CreateGameNote_Success(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"My Note"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Note{ID: uuid.New(), Title: "My Note", GameID: gameID, UserID: authUserID, Visibility: "visible"}
	mockSvc.On("CreateNote", gameID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateGameNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_CreateGameNote_MissingTitle(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	err := h.CreateGameNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestNoteHandler_ListGameNotes_Success(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	notes := []models.Note{{ID: uuid.New(), Title: "Note 1"}}
	mockSvc.On("ListGameNotes", gameID, authUserID, mock.AnythingOfType("repositories.NoteFilters")).Return(notes, nil)

	err := h.ListGameNotes(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_ListGameNotes_WithSessionFilter(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()
	sessionID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?session_id="+sessionID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	notes := []models.Note{}
	mockSvc.On("ListGameNotes", gameID, authUserID, mock.MatchedBy(func(f repositories.NoteFilters) bool {
		return f.SessionID != nil && *f.SessionID == sessionID
	})).Return(notes, nil)

	err := h.ListGameNotes(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_ListGameNotes_InvalidSessionID(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/?session_id=not-a-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.ListGameNotes(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNoteHandler_GetNote_Success(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Note{ID: id, Title: "Note"}
	mockSvc.On("GetNote", id, authUserID).Return(expected, nil)

	err := h.GetNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_GetNote_NotFound(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetNote", id, authUserID).Return(models.Note{}, gorm.ErrRecordNotFound)

	err := h.GetNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_UpdateNote_Success(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	updated := models.Note{ID: id, Title: "Updated", GameID: gameID, UserID: authUserID, Visibility: "visible"}
	mockSvc.On("UpdateNote", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_Success(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	note := models.Note{ID: id, GameID: gameID, UserID: authUserID, Visibility: "visible"}
	mockSvc.On("GetNote", id, authUserID).Return(note, nil)
	mockSvc.On("DeleteNote", id, authUserID).Return(nil)

	err := h.DeleteNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetNote", id, authUserID).Return(models.Note{}, services.ErrForbidden)

	err := h.DeleteNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_CreateGameNote_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"My Note"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	var note models.Note
	note.Title = "My Note"
	note.UserID = authUserID
	mockSvc.On("CreateNote", gameID, authUserID, mock.Anything).Return(models.Note{}, services.ErrForbidden)

	err := h.CreateGameNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_CreateGameNote_InternalError(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"title":"My Note"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateNote", gameID, authUserID, mock.Anything).Return(models.Note{}, errors.New("db error"))

	err := h.CreateGameNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_ListGameNotes_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameNotes", gameID, authUserID, mock.AnythingOfType("repositories.NoteFilters")).Return([]models.Note(nil), services.ErrForbidden)

	err := h.ListGameNotes(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_ListGameNotes_InternalError(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameNotes", gameID, authUserID, mock.AnythingOfType("repositories.NoteFilters")).Return([]models.Note(nil), errors.New("db error"))

	err := h.ListGameNotes(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_GetNote_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetNote", id, authUserID).Return(models.Note{}, services.ErrForbidden)

	err := h.GetNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_GetNote_InternalError(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetNote", id, authUserID).Return(models.Note{}, errors.New("db error"))

	err := h.GetNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_UpdateNote_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	mockSvc.On("UpdateNote", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Note{}, services.ErrForbidden)

	err := h.UpdateNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_UpdateNote_Validation(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"visibility":"bad"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateNote", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Note{}, services.ErrValidation)

	err := h.UpdateNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_UpdateNote_NotFound(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	mockSvc.On("UpdateNote", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Note{}, gorm.ErrRecordNotFound)

	err := h.UpdateNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_UpdateNote_InternalError(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	mockSvc.On("UpdateNote", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Note{}, errors.New("db error"))

	err := h.UpdateNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_NotFound(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetNote", id, authUserID).Return(models.Note{}, gorm.ErrRecordNotFound)

	err := h.DeleteNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_GetNoteInternalError(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetNote", id, authUserID).Return(models.Note{}, errors.New("db error"))

	err := h.DeleteNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_DeleteForbidden(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	note := models.Note{ID: id, GameID: gameID, UserID: authUserID, Visibility: "visible"}
	mockSvc.On("GetNote", id, authUserID).Return(note, nil)
	mockSvc.On("DeleteNote", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestNoteHandler_DeleteNote_DeleteInternalError(t *testing.T) {
	mockSvc := &mocks.MockNoteService{}
	h := newNoteHandler(mockSvc)
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

	note := models.Note{ID: id, GameID: gameID, UserID: authUserID, Visibility: "visible"}
	mockSvc.On("GetNote", id, authUserID).Return(note, nil)
	mockSvc.On("DeleteNote", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteNote(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
