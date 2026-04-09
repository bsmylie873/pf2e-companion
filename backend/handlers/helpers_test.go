package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authmw "pf2e-companion/backend/middleware"
)

func TestParseUUID_Valid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	id := uuid.New()
	c.SetParamNames("id")
	c.SetParamValues(id.String())

	result, err := ParseUUID(c, "id")
	require.NoError(t, err)
	assert.Equal(t, id, result)
}

func TestParseUUID_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")

	_, err := ParseUUID(c, "id")
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestValidateRequired_AllPresent(t *testing.T) {
	missing := ValidateRequired(map[string]interface{}{
		"title": "Test",
		"name":  "Alice",
	})
	assert.Empty(t, missing)
}

func TestValidateRequired_MissingField(t *testing.T) {
	missing := ValidateRequired(map[string]interface{}{
		"title": "",
		"name":  "Alice",
	})
	assert.Contains(t, missing, "title")
	assert.NotContains(t, missing, "name")
}

func TestValidateRequired_NilField(t *testing.T) {
	missing := ValidateRequired(map[string]interface{}{
		"field": nil,
	})
	assert.Contains(t, missing, "field")
}

func TestGetAuthUserID_Present(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	id := uuid.New()
	c.Set(authmw.AuthUserIDKey, id)

	result, err := GetAuthUserID(c)
	require.NoError(t, err)
	assert.Equal(t, id, result)
}

func TestGetAuthUserID_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := GetAuthUserID(c)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetAuthUserID_WrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(authmw.AuthUserIDKey, "not-a-uuid-object")

	_, err := GetAuthUserID(c)
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSuccessResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := SuccessResponse(c, http.StatusOK, map[string]string{"key": "value"})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
}

func TestErrorResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ErrorResponse(c, http.StatusNotFound, "not found")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "not found", resp["message"])
}
