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
	hub     *GameEventHub
}

// NewPinHandler constructs a PinHandler with the given service and hub.
func NewPinHandler(service services.PinService, hub *GameEventHub) *PinHandler {
	return &PinHandler{service: service, hub: hub}
}

// pinBroadcastFilter mirrors the authorization rule applied in
// services.PinService.ListMapPins / ListGamePins. Today that rule is
// "user is a member of the game" — a precondition for being registered
// in the hub at all — so the predicate always returns true. If/when
// session-level "details restricted" visibility is added, this predicate
// is the single place to extend it (using the isGM flag captured at
// WebSocket registration time).
func pinBroadcastFilter() BroadcastFilter {
	return func(_ uuid.UUID, _ bool) bool { return true }
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
	if len(pin.Label) > 100 {
		return ErrorResponse(c, http.StatusBadRequest, "label must be 100 characters or fewer")
	}
	if pin.Description != nil && len(*pin.Description) > 1000 {
		return ErrorResponse(c, http.StatusBadRequest, "description must be 1000 characters or fewer")
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

	h.hub.BroadcastExceptFiltered(resp.GameID, authUserID, pinBroadcastFilter(), GameEvent{Type: "pin_created", GameID: resp.GameID, Data: resp})
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
	if len(pin.Label) > 100 {
		return ErrorResponse(c, http.StatusBadRequest, "label must be 100 characters or fewer")
	}
	if pin.Description != nil && len(*pin.Description) > 1000 {
		return ErrorResponse(c, http.StatusBadRequest, "description must be 1000 characters or fewer")
	}
	resp, err := h.service.CreateGamePin(gameID, authUserID, &pin)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create pin")
	}
	h.hub.BroadcastExceptFiltered(resp.GameID, authUserID, pinBroadcastFilter(), GameEvent{Type: "pin_created", GameID: resp.GameID, Data: resp})
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
	if sessionIDVal, ok := updates["session_id"]; ok && sessionIDVal != nil {
		sessionIDStr, ok := sessionIDVal.(string)
		if !ok {
			return ErrorResponse(c, http.StatusBadRequest, "session_id must be a UUID string or null")
		}
		if _, err := uuid.Parse(sessionIDStr); err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid session_id UUID")
		}
	}
	if labelVal, ok := updates["label"]; ok {
		if label, ok := labelVal.(string); ok {
			if len(label) > 100 {
				return ErrorResponse(c, http.StatusBadRequest, "label must be 100 characters or fewer")
			}
		}
	}
	if descVal, ok := updates["description"]; ok && descVal != nil {
		if desc, ok := descVal.(string); ok {
			if len(desc) > 1000 {
				return ErrorResponse(c, http.StatusBadRequest, "description must be 1000 characters or fewer")
			}
		}
	}

	pin, err := h.service.UpdatePin(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrGroupedPinMove) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		}
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update pin")
	}

	h.hub.BroadcastExceptFiltered(pin.GameID, authUserID, pinBroadcastFilter(), GameEvent{Type: "pin_updated", GameID: pin.GameID, Data: pin})
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

	// Fetch pin before deletion to capture gameID for broadcast.
	pin, err := h.service.GetPin(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete pin")
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

	h.hub.BroadcastExceptFiltered(pin.GameID, authUserID, pinBroadcastFilter(), GameEvent{Type: "pin_deleted", GameID: pin.GameID, Data: map[string]interface{}{"id": id}})
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

// CreateMapPin handles POST /games/:id/maps/:mapId/pins.
func (h *PinHandler) CreateMapPin(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	mapID, err := ParseUUID(c, "mapId")
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
	if len(pin.Label) > 100 {
		return ErrorResponse(c, http.StatusBadRequest, "label must be 100 characters or fewer")
	}
	if pin.Description != nil && len(*pin.Description) > 1000 {
		return ErrorResponse(c, http.StatusBadRequest, "description must be 1000 characters or fewer")
	}
	resp, err := h.service.CreateMapPin(mapID, authUserID, &pin)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "map not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create pin")
	}
	h.hub.BroadcastExceptFiltered(resp.GameID, authUserID, pinBroadcastFilter(), GameEvent{Type: "pin_created", GameID: resp.GameID, Data: resp})
	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListMapPins handles GET /games/:id/maps/:mapId/pins.
func (h *PinHandler) ListMapPins(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	mapID, err := ParseUUID(c, "mapId")
	if err != nil {
		return nil
	}
	pins, err := h.service.ListMapPins(mapID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "map not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list pins")
	}
	return SuccessResponse(c, http.StatusOK, pins)
}

// RegisterPinRoutes registers all pin-related routes on the group.
func RegisterPinRoutes(g *echo.Group, service services.PinService, hub *GameEventHub) {
	h := NewPinHandler(service, hub)
	g.POST("/sessions/:id/pins", h.CreatePin)
	g.POST("/games/:id/pins", h.CreateGamePin)
	g.GET("/games/:id/pins", h.ListGamePins)
	g.GET("/pins/:id", h.GetPin)
	g.PATCH("/pins/:id", h.UpdatePin)
	g.DELETE("/pins/:id", h.DeletePin)
	g.POST("/games/:id/maps/:mapId/pins", h.CreateMapPin)
	g.GET("/games/:id/maps/:mapId/pins", h.ListMapPins)
}
