package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	authmw "pf2e-companion/backend/middleware"
)

// SuccessResponse wraps data in a standard {"data": ...} envelope and sends the response.
func SuccessResponse(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, map[string]interface{}{
		"data": data,
	})
}

// ErrorResponse sends a standard {"code": ..., "message": "..."} error envelope.
func ErrorResponse(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]interface{}{
		"code":    status,
		"message": message,
	})
}

// ParseUUID parses a UUID path parameter by name.
// On failure it writes a 400 Bad Request response and returns a non-nil error.
func ParseUUID(c echo.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		_ = ErrorResponse(c, http.StatusBadRequest, "invalid UUID: "+raw)
		return uuid.Nil, err
	}
	return id, nil
}

// ValidateRequired checks that none of the provided fields are nil or zero-value strings.
// It returns a slice of field names that are missing / empty.
func ValidateRequired(fields map[string]interface{}) []string {
	var missing []string
	for name, val := range fields {
		switch v := val.(type) {
		case string:
			if v == "" {
				missing = append(missing, name)
			}
		case nil:
			missing = append(missing, name)
		}
	}
	return missing
}

// GetAuthUserID retrieves the authenticated user's UUID from the Echo context.
// Returns an error and writes a 401 response if not set.
func GetAuthUserID(c echo.Context) (uuid.UUID, error) {
	val := c.Get(authmw.AuthUserIDKey)
	if val == nil {
		_ = ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, errors.New("unauthorized")
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		_ = ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, errors.New("unauthorized")
	}
	return id, nil
}
