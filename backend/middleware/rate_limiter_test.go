package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_ReturnsMiddlewareFunc(t *testing.T) {
	mw := RateLimiter()
	assert.NotNil(t, mw)
}

func TestBackupRateLimiter_ReturnsMiddlewareFunc(t *testing.T) {
	mw := BackupRateLimiter()
	assert.NotNil(t, mw)
}

func TestPasswordResetRateLimiter_ReturnsMiddlewareFunc(t *testing.T) {
	mw := PasswordResetRateLimiter()
	assert.NotNil(t, mw)
}

func TestRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	e := echo.New()
	mw := RateLimiter()

	callCount := 0
	handler := mw(func(c echo.Context) error {
		callCount++
		return c.JSON(http.StatusOK, nil)
	})

	// Send a few requests — they should all pass since limit is 60/min with burst=10
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
	assert.Equal(t, 5, callCount)
}

func TestRateLimiter_BlocksAfterBurst(t *testing.T) {
	e := echo.New()
	mw := RateLimiter()

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	// Use a unique IP to avoid interference from other tests
	ip := "10.99.99.1:9999"

	// The burst is 10, so 11th request from same IP should be rate limited
	// (rate is 60/min = 1/s, but burst allows 10 at once)
	results := make([]int, 15)
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		if err != nil {
			results[i] = 429 // rate limit error
		} else {
			results[i] = rec.Code
		}
	}

	// At least some requests should have been blocked (429 or 403)
	blocked := 0
	for _, code := range results {
		if code == http.StatusTooManyRequests || code == http.StatusForbidden {
			blocked++
		}
	}
	assert.Greater(t, blocked, 0, "expected some requests to be rate limited")
}

func TestBackupRateLimiter_AllowsInitialRequests(t *testing.T) {
	e := echo.New()
	mw := BackupRateLimiter()

	handler := mw(func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	})

	// Backup limiter has burst=10, so first requests should go through
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/backup", nil)
		req.RemoteAddr = "192.168.1.1:8080"
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set(AuthUserIDKey, nil) // no user

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}
