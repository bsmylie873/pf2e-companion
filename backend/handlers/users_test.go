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

func TestUserHandler_ListUsers_Success(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	users := []models.UserPublicResponse{{ID: uuid.New(), Username: "alice"}}
	mockSvc.On("ListUsers").Return(users, nil)

	err := h.ListUsers(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetUser_Self(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(authUserID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.UserResponse{ID: authUserID, Username: "alice", Email: "alice@test.com"}
	mockSvc.On("GetUser", authUserID).Return(resp, nil)

	err := h.GetUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetUser_Other(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	otherID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(otherID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.UserResponse{ID: otherID, Username: "bob", Email: "bob@test.com"}
	mockSvc.On("GetUser", otherID).Return(resp, nil)

	err := h.GetUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetUser", id).Return(models.UserResponse{}, gorm.ErrRecordNotFound)

	err := h.GetUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_Success(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	body := `{"username":"new_alice"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(authUserID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	resp := models.UserResponse{ID: authUserID, Username: "new_alice"}
	mockSvc.On("UpdateUser", authUserID, mock.AnythingOfType("map[string]interface {}"), authUserID).Return(resp, nil)

	err := h.UpdateUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_ForbiddenOtherUser(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	otherID := uuid.New()

	body := `{"username":"hacker"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(otherID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	err := h.UpdateUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(authUserID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteUser", authUserID, authUserID).Return(nil)

	err := h.DeleteUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_Forbidden(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteUser", id, authUserID).Return(services.ErrForbidden)

	err := h.DeleteUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_ListUsers_InternalError(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("ListUsers").Return([]models.UserPublicResponse(nil), errors.New("db error"))

	err := h.ListUsers(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetUser_InternalError(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("GetUser", id).Return(models.UserResponse{}, errors.New("db error"))

	err := h.GetUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_ServiceForbidden(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	body := `{"username":"new_alice"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(authUserID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateUser", authUserID, mock.AnythingOfType("map[string]interface {}"), authUserID).Return(models.UserResponse{}, services.ErrForbidden)

	err := h.UpdateUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_NotFound(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	body := `{"username":"new_alice"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(authUserID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateUser", authUserID, mock.AnythingOfType("map[string]interface {}"), authUserID).Return(models.UserResponse{}, gorm.ErrRecordNotFound)

	err := h.UpdateUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_InternalError(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()

	body := `{"username":"new_alice"}`
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(authUserID.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("UpdateUser", authUserID, mock.AnythingOfType("map[string]interface {}"), authUserID).Return(models.UserResponse{}, errors.New("db error"))

	err := h.UpdateUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_NotFound(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteUser", id, authUserID).Return(gorm.ErrRecordNotFound)

	err := h.DeleteUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_InternalError(t *testing.T) {
	mockSvc := &mocks.MockUserService{}
	h := NewUserHandler(mockSvc)
	e := echo.New()
	authUserID := uuid.New()
	id := uuid.New()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	c.Set(authmw.AuthUserIDKey, authUserID)

	mockSvc.On("DeleteUser", id, authUserID).Return(errors.New("db error"))

	err := h.DeleteUser(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockSvc.AssertExpectations(t)
}
