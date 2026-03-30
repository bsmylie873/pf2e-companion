package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"pf2e-companion/backend/services"
)

// PreferenceHandler holds the service dependency for preference routes.
type PreferenceHandler struct {
	service services.PreferenceService
}

// NewPreferenceHandler constructs a PreferenceHandler with the given service.
func NewPreferenceHandler(service services.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{service: service}
}

// GetPreferences handles GET /preferences.
func (h *PreferenceHandler) GetPreferences(c echo.Context) error {
	userID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	resp, err := h.service.GetPreferences(userID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch preferences")
	}

	return SuccessResponse(c, http.StatusOK, resp)
}

// UpdatePreferences handles PATCH /preferences.
func (h *PreferenceHandler) UpdatePreferences(c echo.Context) error {
	userID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.UpdatePreferences(userID, updates)
	if err != nil {
		if errors.Is(err, services.ErrValidation) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update preferences")
	}

	return SuccessResponse(c, http.StatusOK, resp)
}

// RegisterPreferenceRoutes wires up all preference routes on the group.
func RegisterPreferenceRoutes(g *echo.Group, service services.PreferenceService) {
	h := NewPreferenceHandler(service)
	g.GET("/preferences", h.GetPreferences)
	g.PATCH("/preferences", h.UpdatePreferences)
}
