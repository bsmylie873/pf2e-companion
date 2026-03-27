package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

// PinHandler holds the service dependency for pin-related routes.
type PinHandler struct {
	service services.PinService
}

// NewPinHandler constructs a PinHandler with the given service.
func NewPinHandler(service services.PinService) *PinHandler {
	return &PinHandler{service: service}
}

// CreatePin handles POST /sessions/:id/pins.
func (h *PinHandler) CreatePin(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	sessionID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var pin models.SessionPin
	if err := c.Bind(&pin); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if pin.Colour == "" {
		pin.Colour = "grey"
	}
	if pin.Icon == "" {
		pin.Icon = "position-marker"
	}
	if err := ValidatePinColour(pin.Colour); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	if err := ValidatePinIcon(pin.Icon); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	resp, err := h.service.CreatePin(sessionID, authUserID, &pin)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create pin")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// CreateGamePin handles POST /games/:id/pins.
func (h *PinHandler) CreateGamePin(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var pin models.SessionPin
	if err := c.Bind(&pin); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if pin.Colour == "" {
		pin.Colour = "grey"
	}
	if pin.Icon == "" {
		pin.Icon = "position-marker"
	}
	if err := ValidatePinColour(pin.Colour); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	if err := ValidatePinIcon(pin.Icon); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	resp, err := h.service.CreateGamePin(gameID, authUserID, &pin)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create pin")
	}
	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGamePins handles GET /games/:id/pins.
func (h *PinHandler) ListGamePins(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	pins, err := h.service.ListGamePins(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list pins")
	}

	return SuccessResponse(c, http.StatusOK, pins)
}

// GetPin handles GET /pins/:id.
func (h *PinHandler) GetPin(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	pin, err := h.service.GetPin(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch pin")
	}

	return SuccessResponse(c, http.StatusOK, pin)
}

// UpdatePin handles PATCH /pins/:id.
func (h *PinHandler) UpdatePin(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if colourVal, ok := updates["colour"]; ok {
		colour, ok := colourVal.(string)
		if !ok {
			return ErrorResponse(c, http.StatusBadRequest, "colour must be a string")
		}
		if err := ValidatePinColour(colour); err != nil {
			return ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
	}
	if iconVal, ok := updates["icon"]; ok {
		icon, ok := iconVal.(string)
		if !ok {
			return ErrorResponse(c, http.StatusBadRequest, "icon must be a string")
		}
		if err := ValidatePinIcon(icon); err != nil {
			return ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
	}

	if noteIDVal, ok := updates["note_id"]; ok && noteIDVal != nil {
		noteIDStr, ok := noteIDVal.(string)
		if !ok {
			return ErrorResponse(c, http.StatusBadRequest, "note_id must be a UUID string or null")
		}
		if _, err := uuid.Parse(noteIDStr); err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid note_id UUID")
		}
	}

	pin, err := h.service.UpdatePin(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update pin")
	}

	return SuccessResponse(c, http.StatusOK, pin)
}

// DeletePin handles DELETE /pins/:id.
func (h *PinHandler) DeletePin(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeletePin(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete pin")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

// RegisterPinRoutes registers all pin-related routes on the group.
func RegisterPinRoutes(g *echo.Group, service services.PinService) {
	h := NewPinHandler(service)
	g.POST("/sessions/:id/pins", h.CreatePin)
	g.POST("/games/:id/pins", h.CreateGamePin)
	g.GET("/games/:id/pins", h.ListGamePins)
	g.GET("/pins/:id", h.GetPin)
	g.PATCH("/pins/:id", h.UpdatePin)
	g.DELETE("/pins/:id", h.DeletePin)
}
