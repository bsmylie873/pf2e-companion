package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"pf2e-companion/backend/services"
)

// PartyMarkerHandler holds the service dependency for party-marker routes.
type PartyMarkerHandler struct {
	service services.PartyMarkerService
	hub     *GameEventHub
}

// RegisterPartyMarkerRoutes registers GET/PUT/DELETE for the single party marker.
func RegisterPartyMarkerRoutes(g *echo.Group, svc services.PartyMarkerService, hub *GameEventHub) {
	h := &PartyMarkerHandler{service: svc, hub: hub}
	g.GET("/games/:id/party-marker", h.GetPartyMarker)
	g.PUT("/games/:id/party-marker", h.UpsertPartyMarker)
	g.DELETE("/games/:id/party-marker", h.DeletePartyMarker)
}

// GetPartyMarker handles GET /games/:id/party-marker.
func (h *PartyMarkerHandler) GetPartyMarker(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	marker, err := h.service.GetPartyMarker(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch party marker")
	}
	return SuccessResponse(c, http.StatusOK, marker)
}

type upsertPartyMarkerRequest struct {
	MapID string  `json:"map_id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

// UpsertPartyMarker handles PUT /games/:id/party-marker.
func (h *PartyMarkerHandler) UpsertPartyMarker(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var req upsertPartyMarkerRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	mapID, err := uuid.Parse(req.MapID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid map_id UUID")
	}
	// Check pre-existence to determine 201 vs 200.
	existing, err := h.service.GetPartyMarker(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch party marker")
	}
	marker, err := h.service.UpsertPartyMarker(gameID, authUserID, mapID, req.X, req.Y)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, services.ErrValidation) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to upsert party marker")
	}
	h.hub.BroadcastExcept(gameID, authUserID, GameEvent{Type: "party_marker_updated", GameID: gameID, Data: marker})
	statusCode := http.StatusOK
	if existing == nil {
		statusCode = http.StatusCreated
	}
	return SuccessResponse(c, statusCode, marker)
}

// DeletePartyMarker handles DELETE /games/:id/party-marker.
func (h *PartyMarkerHandler) DeletePartyMarker(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	if err := h.service.DeletePartyMarker(gameID, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete party marker")
	}
	h.hub.BroadcastExcept(gameID, authUserID, GameEvent{Type: "party_marker_deleted", GameID: gameID, Data: map[string]interface{}{"game_id": gameID}})
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}
