package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRF_ReturnsMiddlewareFunc(t *testing.T) {
	// Verify that CSRF() returns a non-nil middleware function
	mw := CSRF()
	assert.NotNil(t, mw)
}

func TestCSRF_SkipsAuthRoutes(t *testing.T) {
	// The CSRF middleware is configured to skip /auth/* routes.
	// We test that requests to /auth/login pass through without a CSRF token.
	e := echo.New()
	mw := CSRF()

	called := false
	handler := mw(func(c echo.Context) error {
		called = true
		return c.JSON(http.StatusOK, nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/auth/login")

	err := handler(c)
	require.NoError(t, err)
	assert.True(t, called, "handler should have been called for /auth/ routes")
}

func TestCSRF_NonAuthRoute_NoCsrfToken_Blocked(t *testing.T) {
	// For non-auth routes, a missing CSRF token header should result in a 403.
	e := echo.New()
	mw := CSRF()

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/games")

	// Without a csrf_token cookie, the CSRF middleware should block
	// (it needs the cookie to verify against the header)
	err := handler(c)
	// The CSRF middleware either returns an error or writes 403 itself
	if err != nil {
		// Some versions return an error that echo would handle
		assert.NotNil(t, err)
	} else {
		// The middleware wrote the response directly
		assert.Equal(t, http.StatusForbidden, rec.Code)
	}
}
