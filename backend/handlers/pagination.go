package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PaginatedResponse is the envelope returned for paginated list endpoints.
// It intentionally mirrors the {"data": ...} envelope so it can be returned directly via c.JSON.
type PaginatedResponse struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// ParsePaginationParams reads optional "page" and "limit" query parameters.
// Returns (page, limit, paginated=true, nil) when the "page" param is present and valid.
// Returns (0, 0, false, nil) when both "page" and "limit" are absent — caller should use the non-paginated path.
// On validation failure it writes a 400 response and returns an error.
func ParsePaginationParams(c echo.Context) (page int, limit int, paginated bool, err error) {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	if pageStr == "" && limitStr == "" {
		return 0, 0, false, nil
	}

	page = 1
	limit = DefaultPageSize

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			_ = ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("invalid page parameter: %q must be a positive integer", pageStr))
			return 0, 0, false, fmt.Errorf("invalid page")
		}
		if page < 1 {
			_ = ErrorResponse(c, http.StatusBadRequest, "invalid page parameter: must be >= 1")
			return 0, 0, false, fmt.Errorf("page < 1")
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			_ = ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("invalid limit parameter: %q must be a positive integer", limitStr))
			return 0, 0, false, fmt.Errorf("invalid limit")
		}
		if limit < 1 || limit > MaxPageSize {
			_ = ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("invalid limit parameter: must be between 1 and %d", MaxPageSize))
			return 0, 0, false, fmt.Errorf("limit out of range")
		}
	}

	return page, limit, true, nil
}
