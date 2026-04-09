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
	"github.com/stretchr/testify/require"
	authmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

func TestInviteHandler_GenerateInvite_Success(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"expires_in":"7d"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.InviteTokenResponse{Token: "invite123"}
	mockSvc.On("GenerateInvite", gameID, authUserID, "7d").Return(resp, nil)

	err := h.GenerateInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GenerateInvite_DefaultNever(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
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

	resp := models.InviteTokenResponse{Token: "invite123"}
	mockSvc.On("GenerateInvite", gameID, authUserID, "never").Return(resp, nil)

	err := h.GenerateInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GenerateInvite_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"expires_in":"24h"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GenerateInvite", gameID, authUserID, "24h").Return(models.InviteTokenResponse{}, services.ErrForbidden)

	err := h.GenerateInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GetActiveInvite_Success(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.InviteTokenStatusResponse{HasActiveInvite: true}
	mockSvc.On("GetActiveInvite", gameID, authUserID).Return(resp, nil)

	err := h.GetActiveInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_RevokeInvite_Success(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RevokeInvite", gameID, authUserID).Return(nil)

	err := h.RevokeInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestValidateInvite_Success(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("raw_invite_token")

	resp := models.InviteValidationResponse{GameTitle: "My Campaign"}
	mockSvc.On("ValidateInvite", "raw_invite_token").Return(resp, nil)

	handler := ValidateInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestValidateInvite_Invalid(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("bad_token")

	mockSvc.On("ValidateInvite", "bad_token").Return(models.InviteValidationResponse{}, errors.New("invalid token"))

	handler := ValidateInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestRedeemInvite_Success(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("raw_invite_token")
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.InviteRedeemResponse{GameID: uuid.New()}
	mockSvc.On("RedeemInvite", "raw_invite_token", authUserID).Return(resp, nil)

	handler := RedeemInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestRedeemInvite_Expired(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("expired_token")
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RedeemInvite", "expired_token", authUserID).Return(models.InviteRedeemResponse{}, errors.New("expired invite"))

	handler := RedeemInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GenerateInvite_Validation(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"expires_in":"bad_value"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GenerateInvite", gameID, authUserID, "bad_value").Return(models.InviteTokenResponse{}, services.ErrValidation)

	err := h.GenerateInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GenerateInvite_InternalError(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	body := `{"expires_in":"7d"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GenerateInvite", gameID, authUserID, "7d").Return(models.InviteTokenResponse{}, errors.New("db error"))

	err := h.GenerateInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GetActiveInvite_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetActiveInvite", gameID, authUserID).Return(models.InviteTokenStatusResponse{}, services.ErrForbidden)

	err := h.GetActiveInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_GetActiveInvite_InternalError(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetActiveInvite", gameID, authUserID).Return(models.InviteTokenStatusResponse{}, errors.New("db error"))

	err := h.GetActiveInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_RevokeInvite_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RevokeInvite", gameID, authUserID).Return(services.ErrForbidden)

	err := h.RevokeInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestInviteHandler_RevokeInvite_InternalError(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	h := NewInviteHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	gameID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(gameID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RevokeInvite", gameID, authUserID).Return(errors.New("db error"))

	err := h.RevokeInvite(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestValidateInvite_EmptyToken(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// no param "token" set — c.Param("token") returns ""

	handler := ValidateInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestValidateInvite_InternalError(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("some_token")

	mockSvc.On("ValidateInvite", "some_token").Return(models.InviteValidationResponse{}, errors.New("db error"))

	handler := ValidateInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestRedeemInvite_EmptyToken(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// no param "token" set — c.Param("token") returns ""
	c.Set(authmw.AuthUserIDKey, authUserID)

	handler := RedeemInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRedeemInvite_InternalError(t *testing.T) {
	mockSvc := &mocks.MockInviteService{}
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("some_token")
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("RedeemInvite", "some_token", authUserID).Return(models.InviteRedeemResponse{}, errors.New("db error"))

	handler := RedeemInvite(mockSvc)
	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
