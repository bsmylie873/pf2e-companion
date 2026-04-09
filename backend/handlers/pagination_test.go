package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePaginationParams_NoneProvided(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, limit, paginated, err := ParsePaginationParams(c)
	require.NoError(t, err)
	assert.False(t, paginated)
	assert.Equal(t, 0, page)
	assert.Equal(t, 0, limit)
}

func TestParsePaginationParams_PageOnly(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?page=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, limit, paginated, err := ParsePaginationParams(c)
	require.NoError(t, err)
	assert.True(t, paginated)
	assert.Equal(t, 2, page)
	assert.Equal(t, DefaultPageSize, limit)
}

func TestParsePaginationParams_PageAndLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?page=3&limit=25", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	page, limit, paginated, err := ParsePaginationParams(c)
	require.NoError(t, err)
	assert.True(t, paginated)
	assert.Equal(t, 3, page)
	assert.Equal(t, 25, limit)
}

func TestParsePaginationParams_InvalidPage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, _, _, err := ParsePaginationParams(c)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestParsePaginationParams_PageZero(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?page=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, _, _, err := ParsePaginationParams(c)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestParsePaginationParams_LimitTooLarge(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=200", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, _, _, err := ParsePaginationParams(c)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestParsePaginationParams_InvalidLimit(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?limit=xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, _, _, err := ParsePaginationParams(c)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
