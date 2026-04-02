package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

// SessionHandler holds the service dependency for session-related routes.
type SessionHandler struct {
	service services.SessionService
	hub     *GameEventHub
}

// NewSessionHandler constructs a SessionHandler with the given service and hub.
func NewSessionHandler(service services.SessionService, hub *GameEventHub) *SessionHandler {
	return &SessionHandler{service: service, hub: hub}
}

// CreateSession handles POST /games/:id/sessions.
func (h *SessionHandler) CreateSession(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var session models.Session
	if err := c.Bind(&session); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{
		"title": session.Title,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	if session.RuntimeStart != nil && session.RuntimeEnd != nil && !session.RuntimeEnd.After(*session.RuntimeStart) {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "runtime_end must be after runtime_start")
	}

	resp, err := h.service.CreateSession(gameID, authUserID, &session)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create session")
	}

	h.hub.BroadcastExcept(gameID, authUserID, GameEvent{Type: "session_created", GameID: gameID, Data: resp})
	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameSessions handles GET /games/:id/sessions.
func (h *SessionHandler) ListGameSessions(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	sessions, err := h.service.ListGameSessions(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list sessions")
	}

	return SuccessResponse(c, http.StatusOK, sessions)
}

// GetSession handles GET /sessions/:id.
func (h *SessionHandler) GetSession(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	session, err := h.service.GetSession(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch session")
	}

	return SuccessResponse(c, http.StatusOK, session)
}

// UpdateSession handles PATCH /sessions/:id.
func (h *SessionHandler) UpdateSession(c echo.Context) error {
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

	// Validate runtime_end is after runtime_start when both are present in the update.
	if startStr, hasStart := updates["runtime_start"]; hasStart {
		if endStr, hasEnd := updates["runtime_end"]; hasEnd {
			if startStr != nil && endStr != nil {
				start, errS := time.Parse(time.RFC3339, startStr.(string))
				end, errE := time.Parse(time.RFC3339, endStr.(string))
				if errS == nil && errE == nil && !end.After(start) {
					return ErrorResponse(c, http.StatusUnprocessableEntity, "runtime_end must be after runtime_start")
				}
			}
		}
	}

	session, err := h.service.UpdateSession(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update session")
	}

	h.hub.BroadcastExcept(session.GameID, authUserID, GameEvent{Type: "session_updated", GameID: session.GameID, Data: session})
	return SuccessResponse(c, http.StatusOK, session)
}

// DeleteSession handles DELETE /sessions/:id.
func (h *SessionHandler) DeleteSession(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	// Fetch session before deletion to capture gameID for broadcast.
	session, err := h.service.GetSession(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete session")
	}

	if err := h.service.DeleteSession(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete session")
	}

	h.hub.BroadcastExcept(session.GameID, authUserID, GameEvent{Type: "session_deleted", GameID: session.GameID, Data: map[string]interface{}{"id": id}})
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

// RegisterSessionRoutes registers all session-related routes on the group.
func RegisterSessionRoutes(g *echo.Group, service services.SessionService, hub *GameEventHub) {
	h := NewSessionHandler(service, hub)
	g.POST("/games/:id/sessions", h.CreateSession)
	g.GET("/games/:id/sessions", h.ListGameSessions)
	g.GET("/sessions/:id", h.GetSession)
	g.PATCH("/sessions/:id", h.UpdateSession)
	g.DELETE("/sessions/:id", h.DeleteSession)
}
