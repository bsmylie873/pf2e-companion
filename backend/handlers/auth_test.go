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
	"github.com/stretchr/testify/require"
	authmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice","email":"alice@test.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user := models.UserResponse{ID: uuid.New(), Username: "alice"}
	pair := models.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	mockSvc.On("Register", models.RegisterRequest{Username: "alice", Email: "alice@test.com", Password: "secret"}).Return(user, pair, nil)

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_MissingFields(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAuthHandler_Register_DuplicateUser(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice","email":"alice@test.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Register", models.RegisterRequest{Username: "alice", Email: "alice@test.com", Password: "secret"}).Return(models.UserResponse{}, models.TokenPair{}, errors.New("duplicate key"))

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	user := models.UserResponse{ID: uuid.New(), Username: "alice"}
	pair := models.TokenPair{AccessToken: "access", RefreshToken: "refresh"}
	mockSvc.On("Login", models.LoginRequest{Username: "alice", Password: "secret"}).Return(user, pair, nil)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Login", models.LoginRequest{Username: "alice", Password: "wrong"}).Return(models.UserResponse{}, models.TokenPair{}, errors.New("invalid credentials"))

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAuthHandler_Logout_WithRefreshCookie(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "mytoken"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Logout", "mytoken").Return(nil)

	err := h.Logout(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Logout_WithoutCookie(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Logout(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid_refresh"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	pair := models.TokenPair{AccessToken: "new_access", RefreshToken: "new_refresh"}
	mockSvc.On("RefreshTokens", "valid_refresh").Return(pair, nil)

	err := h.Refresh(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_NoCookie(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Refresh(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	user := models.UserResponse{ID: authUserID, Username: "alice"}
	mockSvc.On("GetMe", authUserID).Return(user, nil)

	err := h.Me(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ForgotPassword_ReturnsOK(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"email":"alice@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("RequestPasswordReset", "alice@test.com").Return("reset_token_123", nil)

	err := h.ForgotPassword(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_Success(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"token":"reset_tok","new_password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("ResetPassword", "reset_tok", "newpass123").Return(nil)

	err := h.ResetPassword(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_MissingFields(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"token":"tok"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ResetPassword(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice","email":"alice@test.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Register", models.RegisterRequest{Username: "alice", Email: "alice@test.com", Password: "secret"}).Return(models.UserResponse{}, models.TokenPair{}, errors.New("db connection failed"))

	err := h.Register(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"username":"alice","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Login", models.LoginRequest{Username: "alice", Password: "secret"}).Return(models.UserResponse{}, models.TokenPair{}, errors.New("db error"))

	err := h.Login(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad_token"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("RefreshTokens", "bad_token").Return(models.TokenPair{}, errors.New("invalid or expired refresh token"))

	err := h.Refresh(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Me_NotFound(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetMe", authUserID).Return(models.UserResponse{}, errors.New("not found"))

	err := h.Me(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ForgotPassword_NoToken(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"email":"nobody@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Email not found — service returns empty string
	mockSvc.On("RequestPasswordReset", "nobody@example.com").Return("", nil)

	err := h.ForgotPassword(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Nil(t, data["token"])
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_InvalidToken(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"token":"bad_tok","new_password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("ResetPassword", "bad_tok", "newpass123").Return(errors.New("invalid or expired token"))

	err := h.ResetPassword(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_ResetPassword_InternalError(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	h := NewAuthHandler(mockSvc)
	e := echo.New()

	body := `{"token":"tok","new_password":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("ResetPassword", "tok", "newpass123").Return(errors.New("db error"))

	err := h.ResetPassword(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
