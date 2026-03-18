package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
