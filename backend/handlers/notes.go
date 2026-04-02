package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/repositories"
	"pf2e-companion/backend/services"
)

// NoteHandler handles HTTP requests for notes.
type NoteHandler struct {
	service services.NoteService
	hub     *GameEventHub
}

// NewNoteHandler creates a new NoteHandler with the given service and hub.
func NewNoteHandler(service services.NoteService, hub *GameEventHub) *NoteHandler {
	return &NoteHandler{service: service, hub: hub}
}

// RegisterNoteRoutes wires all note routes onto the group.
func RegisterNoteRoutes(g *echo.Group, service services.NoteService, hub *GameEventHub) {
	h := NewNoteHandler(service, hub)
	g.POST("/games/:id/notes", h.CreateGameNote)
	g.GET("/games/:id/notes", h.ListGameNotes)
	g.GET("/notes/:id", h.GetNote)
	g.PATCH("/notes/:id", h.UpdateNote)
	g.DELETE("/notes/:id", h.DeleteNote)
}

// CreateGameNote creates a note belonging to a game.
// POST /games/:id/notes
func (h *NoteHandler) CreateGameNote(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var note models.Note
	if err := c.Bind(&note); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{"title": note.Title})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	note.UserID = authUserID

	resp, err := h.service.CreateNote(gameID, authUserID, &note)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create note")
	}

	if resp.Visibility == "private" {
		h.hub.SendToUser(gameID, authUserID, GameEvent{Type: "note_created", GameID: gameID, Data: resp})
	} else {
		h.hub.BroadcastExcept(gameID, authUserID, GameEvent{Type: "note_created", GameID: gameID, Data: resp})
	}
	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameNotes returns notes belonging to a game, filtered by query params.
// GET /games/:id/notes
func (h *NoteHandler) ListGameNotes(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	filters := repositories.NoteFilters{
		Sort:     c.QueryParam("sort"),
		Unlinked: c.QueryParam("unlinked") == "true",
	}

	if sessionIDStr := c.QueryParam("session_id"); sessionIDStr != "" {
		parsed, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid session_id")
		}
		filters.SessionID = &parsed
	}

	notes, err := h.service.ListGameNotes(gameID, authUserID, filters)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch notes")
	}

	return SuccessResponse(c, http.StatusOK, notes)
}

// GetNote retrieves a single note by ID.
// GET /notes/:id
func (h *NoteHandler) GetNote(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	note, err := h.service.GetNote(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch note")
	}

	return SuccessResponse(c, http.StatusOK, note)
}

// UpdateNote applies a partial update to a note.
// PATCH /notes/:id
func (h *NoteHandler) UpdateNote(c echo.Context) error {
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

	note, err := h.service.UpdateNote(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, services.ErrValidation) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update note")
	}

	if note.Visibility == "private" {
		h.hub.SendToUser(note.GameID, note.UserID, GameEvent{Type: "note_updated", GameID: note.GameID, Data: note})
	} else {
		h.hub.BroadcastExcept(note.GameID, authUserID, GameEvent{Type: "note_updated", GameID: note.GameID, Data: note})
	}
	return SuccessResponse(c, http.StatusOK, note)
}

// DeleteNote removes a note by ID.
// DELETE /notes/:id
func (h *NoteHandler) DeleteNote(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	// Fetch note before deletion to capture gameID and visibility for broadcast.
	note, err := h.service.GetNote(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete note")
	}

	if err := h.service.DeleteNote(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete note")
	}

	if note.Visibility == "private" {
		h.hub.SendToUser(note.GameID, note.UserID, GameEvent{Type: "note_deleted", GameID: note.GameID, Data: map[string]interface{}{"id": id}})
	} else {
		h.hub.BroadcastExcept(note.GameID, authUserID, GameEvent{Type: "note_deleted", GameID: note.GameID, Data: map[string]interface{}{"id": id}})
	}
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}
