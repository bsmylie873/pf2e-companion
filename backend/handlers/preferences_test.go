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
	authmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

func TestPreferenceHandler_GetPreferences_Success(t *testing.T) {
	mockSvc := &mocks.MockPreferenceService{}
	h := NewPreferenceHandler(mockSvc)
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, userID)

	pref := models.UserPreferenceResponse{}
	mockSvc.On("GetPreferences", userID).Return(pref, nil)

	err := h.GetPreferences(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPreferenceHandler_GetPreferences_NoAuth(t *testing.T) {
	mockSvc := &mocks.MockPreferenceService{}
	h := NewPreferenceHandler(mockSvc)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetPreferences(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPreferenceHandler_UpdatePreferences_Success(t *testing.T) {
	mockSvc := &mocks.MockPreferenceService{}
	h := NewPreferenceHandler(mockSvc)
	e := echo.New()
	userID := uuid.New()

	body := `{"default_pin_colour":"red"}`
	req := httptest.NewRequest(http.MethodPatch, "/preferences", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, userID)

	pref := models.UserPreferenceResponse{}
	mockSvc.On("UpdatePreferences", userID, mock.AnythingOfType("map[string]interface {}")).Return(pref, nil)

	err := h.UpdatePreferences(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPreferenceHandler_GetPreferences_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPreferenceService{}
	h := NewPreferenceHandler(mockSvc)
	e := echo.New()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, userID)

	mockSvc.On("GetPreferences", userID).Return(models.UserPreferenceResponse{}, errors.New("db error"))

	err := h.GetPreferences(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPreferenceHandler_UpdatePreferences_Validation(t *testing.T) {
	mockSvc := &mocks.MockPreferenceService{}
	h := NewPreferenceHandler(mockSvc)
	e := echo.New()
	userID := uuid.New()

	body := `{"default_pin_colour":"invalid"}`
	req := httptest.NewRequest(http.MethodPatch, "/preferences", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, userID)

	mockSvc.On("UpdatePreferences", userID, mock.AnythingOfType("map[string]interface {}")).Return(models.UserPreferenceResponse{}, services.ErrValidation)

	err := h.UpdatePreferences(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestPreferenceHandler_UpdatePreferences_InternalError(t *testing.T) {
	mockSvc := &mocks.MockPreferenceService{}
	h := NewPreferenceHandler(mockSvc)
	e := echo.New()
	userID := uuid.New()

	body := `{"default_pin_colour":"red"}`
	req := httptest.NewRequest(http.MethodPatch, "/preferences", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, userID)

	mockSvc.On("UpdatePreferences", userID, mock.AnythingOfType("map[string]interface {}")).Return(models.UserPreferenceResponse{}, errors.New("db error"))

	err := h.UpdatePreferences(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
