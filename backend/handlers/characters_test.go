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

func TestCharacterHandler_CreateCharacter_Success(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Aragorn"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Character{ID: uuid.New(), Name: "Aragorn", GameID: gameID}
	mockSvc.On("CreateCharacter", gameID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_CreateCharacter_MissingName(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
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

	err := h.CreateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestCharacterHandler_ListGameCharacters_Success(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	chars := []models.Character{{ID: uuid.New(), Name: "Hero"}}
	mockSvc.On("ListGameCharacters", gameID, authUserID).Return(chars, nil)

	err := h.ListGameCharacters(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_GetCharacter_Success(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Character{ID: id, Name: "Hero"}
	mockSvc.On("GetCharacter", id, authUserID).Return(expected, nil)

	err := h.GetCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_GetCharacter_NotFound(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetCharacter", id, authUserID).Return(models.Character{}, gorm.ErrRecordNotFound)

	err := h.GetCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_UpdateCharacter_Success(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Updated Hero"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.Character{ID: id, Name: "Updated Hero"}
	mockSvc.On("UpdateCharacter", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_UpdateCharacter_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateCharacter", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Character{}, services.ErrForbidden)

	err := h.UpdateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_DeleteCharacter_Success(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteCharacter", id, authUserID).Return(nil)

	err := h.DeleteCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_DeleteCharacter_NotFound(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteCharacter", id, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DeleteCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_CreateCharacter_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Aragorn"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateCharacter", gameID, authUserID, mock.Anything).Return(models.Character{}, services.ErrForbidden)

	err := h.CreateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_CreateCharacter_InternalError(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Aragorn"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateCharacter", gameID, authUserID, mock.Anything).Return(models.Character{}, errors.New("db error"))

	err := h.CreateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_ListGameCharacters_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameCharacters", gameID, authUserID).Return([]models.Character(nil), services.ErrForbidden)

	err := h.ListGameCharacters(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_ListGameCharacters_InternalError(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameCharacters", gameID, authUserID).Return([]models.Character(nil), errors.New("db error"))

	err := h.ListGameCharacters(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_GetCharacter_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetCharacter", id, authUserID).Return(models.Character{}, services.ErrForbidden)

	err := h.GetCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_GetCharacter_InternalError(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetCharacter", id, authUserID).Return(models.Character{}, errors.New("db error"))

	err := h.GetCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_UpdateCharacter_NotFound(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Updated Hero"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateCharacter", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Character{}, gorm.ErrRecordNotFound)

	err := h.UpdateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_UpdateCharacter_InternalError(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Updated Hero"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateCharacter", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Character{}, errors.New("db error"))

	err := h.UpdateCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_DeleteCharacter_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteCharacter", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCharacterHandler_DeleteCharacter_InternalError(t *testing.T) {
	mockSvc := &mocks.MockCharacterService{}
	h := NewCharacterHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteCharacter", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteCharacter(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
