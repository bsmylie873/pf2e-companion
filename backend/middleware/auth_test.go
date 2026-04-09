package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authpkg "pf2e-companion/backend/auth"
	"pf2e-companion/backend/mocks"
	"pf2e-companion/backend/models"
)

func nextHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func TestRequireAuth_ValidAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	mockSvc := &mocks.MockAuthService{}
	e := echo.New()

	userID := uuid.New()
	token, err := authpkg.GenerateAccessToken(userID)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequireAuth(mockSvc)
	err = mw(nextHandler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that the user ID was set in context
	val := c.Get(AuthUserIDKey)
	require.NotNil(t, val)
	assert.Equal(t, userID, val.(uuid.UUID))
}

func TestRequireAuth_NoCookies_Unauthorized(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequireAuth(mockSvc)
	err := mw(nextHandler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuth_InvalidAccessToken_ValidRefreshToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	userID := uuid.New()
	refreshToken, err := authpkg.GenerateRefreshToken(userID)
	require.NoError(t, err)

	newPair := models.TokenPair{
		AccessToken:  func() string { t, _ := authpkg.GenerateAccessToken(userID); return t }(),
		RefreshToken: refreshToken,
	}

	mockSvc := &mocks.MockAuthService{}
	mockSvc.On("RefreshTokens", refreshToken).Return(newPair, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid_access_token"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequireAuth(mockSvc)
	err = mw(nextHandler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestRequireAuth_InvalidAccessToken_InvalidRefreshToken(t *testing.T) {
	mockSvc := &mocks.MockAuthService{}
	mockSvc.On("RefreshTokens", "bad_refresh").Return(models.TokenPair{}, assert.AnError)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "bad_access"})
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad_refresh"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequireAuth(mockSvc)
	err := mw(nextHandler)(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockSvc.AssertExpectations(t)
}

func TestCookieSecure_False(t *testing.T) {
	os.Unsetenv("COOKIE_SECURE")
	assert.False(t, CookieSecure())
}

func TestCookieSecure_True(t *testing.T) {
	os.Setenv("COOKIE_SECURE", "true")
	defer os.Unsetenv("COOKIE_SECURE")
	assert.True(t, CookieSecure())
}

func TestCookieSecure_One(t *testing.T) {
	os.Setenv("COOKIE_SECURE", "1")
	defer os.Unsetenv("COOKIE_SECURE")
	assert.True(t, CookieSecure())
}

func TestSetAccessCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetAccessCookie(c, "myaccesstoken", false)
	cookies := rec.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "access_token" {
			found = true
			assert.Equal(t, "myaccesstoken", cookie.Value)
			assert.True(t, cookie.HttpOnly)
		}
	}
	assert.True(t, found)
}

func TestSetRefreshCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetRefreshCookie(c, "myrefreshtoken", false)
	cookies := rec.Result().Cookies()
	var found bool
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			found = true
			assert.Equal(t, "myrefreshtoken", cookie.Value)
		}
	}
	assert.True(t, found)
}

func TestClearAuthCookies(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ClearAuthCookies(c)

	cookies := rec.Result().Cookies()
	names := make(map[string]*http.Cookie)
	for _, cookie := range cookies {
		names[cookie.Name] = cookie
	}

	require.Contains(t, names, "access_token")
	require.Contains(t, names, "refresh_token")
	require.Contains(t, names, "csrf_token")
	assert.Equal(t, -1, names["access_token"].MaxAge)
	assert.Equal(t, -1, names["refresh_token"].MaxAge)
	assert.Equal(t, -1, names["csrf_token"].MaxAge)
	assert.Equal(t, "", names["access_token"].Value)
	assert.Equal(t, "", names["refresh_token"].Value)
	assert.Equal(t, "", names["csrf_token"].Value)
}
