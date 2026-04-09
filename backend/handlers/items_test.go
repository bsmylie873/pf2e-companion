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

func TestItemHandler_CreateItem_Success(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Sword"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Item{ID: uuid.New(), Name: "Sword", GameID: gameID}
	mockSvc.On("CreateItem", gameID, authUserID, mock.Anything).Return(expected, nil)

	err := h.CreateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_CreateItem_MissingName(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
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

	err := h.CreateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestItemHandler_ListGameItems_Success(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	items := []models.Item{{ID: uuid.New(), Name: "Shield"}}
	mockSvc.On("ListGameItems", gameID, authUserID).Return(items, nil)

	err := h.ListGameItems(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_ListCharacterItems_Success(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	charID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(charID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	items := []models.Item{{ID: uuid.New(), Name: "Helmet"}}
	mockSvc.On("ListCharacterItems", charID, authUserID).Return(items, nil)

	err := h.ListCharacterItems(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_GetItem_Success(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	expected := models.Item{ID: id, Name: "Axe"}
	mockSvc.On("GetItem", id, authUserID).Return(expected, nil)

	err := h.GetItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_GetItem_NotFound(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetItem", id, authUserID).Return(models.Item{}, gorm.ErrRecordNotFound)

	err := h.GetItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_UpdateItem_Success(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Better Sword"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	updated := models.Item{ID: id, Name: "Better Sword"}
	mockSvc.On("UpdateItem", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(updated, nil)

	err := h.UpdateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_UpdateItem_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"X"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateItem", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Item{}, services.ErrForbidden)

	err := h.UpdateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_DeleteItem_Success(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteItem", id, authUserID).Return(nil)

	err := h.DeleteItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_CreateItem_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Sword"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateItem", gameID, authUserID, mock.Anything).Return(models.Item{}, services.ErrForbidden)

	err := h.CreateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_CreateItem_InternalError(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"name":"Sword"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("CreateItem", gameID, authUserID, mock.Anything).Return(models.Item{}, errors.New("db error"))

	err := h.CreateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_ListGameItems_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameItems", gameID, authUserID).Return([]models.Item(nil), services.ErrForbidden)

	err := h.ListGameItems(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_ListGameItems_InternalError(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListGameItems", gameID, authUserID).Return([]models.Item(nil), errors.New("db error"))

	err := h.ListGameItems(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_ListCharacterItems_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	charID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(charID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListCharacterItems", charID, authUserID).Return([]models.Item(nil), services.ErrForbidden)

	err := h.ListCharacterItems(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_ListCharacterItems_InternalError(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	charID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(charID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListCharacterItems", charID, authUserID).Return([]models.Item(nil), errors.New("db error"))

	err := h.ListCharacterItems(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_GetItem_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetItem", id, authUserID).Return(models.Item{}, services.ErrForbidden)

	err := h.GetItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_GetItem_InternalError(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetItem", id, authUserID).Return(models.Item{}, errors.New("db error"))

	err := h.GetItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_UpdateItem_NotFound(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Better Sword"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateItem", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Item{}, gorm.ErrRecordNotFound)

	err := h.UpdateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_UpdateItem_InternalError(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	body := `{"name":"Better Sword"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateItem", id, authUserID, mock.AnythingOfType("map[string]interface {}")).Return(models.Item{}, errors.New("db error"))

	err := h.UpdateItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_DeleteItem_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteItem", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_DeleteItem_NotFound(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteItem", id, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DeleteItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestItemHandler_DeleteItem_InternalError(t *testing.T) {
	mockSvc := &mocks.MockItemService{}
	h := NewItemHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteItem", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteItem(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
