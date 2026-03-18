package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

// SessionHandler holds the service dependency for session-related routes.
type SessionHandler struct {
	service services.SessionService
}

// NewSessionHandler constructs a SessionHandler with the given service.
func NewSessionHandler(service services.SessionService) *SessionHandler {
	return &SessionHandler{service: service}
}

// CreateSession handles POST /games/:id/sessions.
func (h *SessionHandler) CreateSession(c echo.Context) error {
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

	resp, err := h.service.CreateSession(gameID, &session)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create session")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameSessions handles GET /games/:id/sessions.
func (h *SessionHandler) ListGameSessions(c echo.Context) error {
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	sessions, err := h.service.ListGameSessions(gameID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list sessions")
	}

	return SuccessResponse(c, http.StatusOK, sessions)
}

// GetSession handles GET /sessions/:id.
func (h *SessionHandler) GetSession(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	session, err := h.service.GetSession(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch session")
	}

	return SuccessResponse(c, http.StatusOK, session)
}

// UpdateSession handles PATCH /sessions/:id.
func (h *SessionHandler) UpdateSession(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	session, err := h.service.UpdateSession(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update session")
	}

	return SuccessResponse(c, http.StatusOK, session)
}

// DeleteSession handles DELETE /sessions/:id.
func (h *SessionHandler) DeleteSession(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteSession(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete session")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

// RegisterSessionRoutes registers all session-related routes on the Echo instance.
func RegisterSessionRoutes(e *echo.Echo, service services.SessionService) {
	h := NewSessionHandler(service)
	e.POST("/games/:id/sessions", h.CreateSession)
	e.GET("/games/:id/sessions", h.ListGameSessions)
	e.GET("/sessions/:id", h.GetSession)
	e.PATCH("/sessions/:id", h.UpdateSession)
	e.DELETE("/sessions/:id", h.DeleteSession)
}
