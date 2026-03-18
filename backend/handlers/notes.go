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

// NoteHandler handles HTTP requests for notes.
type NoteHandler struct {
	service services.NoteService
}

// NewNoteHandler creates a new NoteHandler with the given service.
func NewNoteHandler(service services.NoteService) *NoteHandler {
	return &NoteHandler{service: service}
}

// RegisterNoteRoutes wires all note routes onto the Echo instance.
func RegisterNoteRoutes(e *echo.Echo, service services.NoteService) {
	h := NewNoteHandler(service)
	e.POST("/games/:id/notes", h.CreateGameNote)
	e.POST("/users/:id/notes", h.CreateUserNote)
	e.GET("/games/:id/notes", h.ListGameNotes)
	e.GET("/users/:id/notes", h.ListUserNotes)
	e.GET("/notes/:id", h.GetNote)
	e.PATCH("/notes/:id", h.UpdateNote)
	e.DELETE("/notes/:id", h.DeleteNote)
}

// CreateGameNote creates a shared note belonging to a game.
// POST /games/:id/notes
func (h *NoteHandler) CreateGameNote(c echo.Context) error {
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

	resp, err := h.service.CreateGameNote(gameID, &note)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create note")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// CreateUserNote creates a private note belonging to a user.
// POST /users/:id/notes
func (h *NoteHandler) CreateUserNote(c echo.Context) error {
	userID, err := ParseUUID(c, "id")
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

	resp, err := h.service.CreateUserNote(userID, &note)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create note")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameNotes returns all notes belonging to a game.
// GET /games/:id/notes
func (h *NoteHandler) ListGameNotes(c echo.Context) error {
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	notes, err := h.service.ListGameNotes(gameID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch notes")
	}

	return SuccessResponse(c, http.StatusOK, notes)
}

// ListUserNotes returns all notes belonging to a user.
// GET /users/:id/notes
func (h *NoteHandler) ListUserNotes(c echo.Context) error {
	userID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	notes, err := h.service.ListUserNotes(userID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch notes")
	}

	return SuccessResponse(c, http.StatusOK, notes)
}

// GetNote retrieves a single note by ID.
// GET /notes/:id
func (h *NoteHandler) GetNote(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	note, err := h.service.GetNote(id)
	if err != nil {
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
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	note, err := h.service.UpdateNote(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update note")
	}

	return SuccessResponse(c, http.StatusOK, note)
}

// DeleteNote removes a note by ID.
// DELETE /notes/:id
func (h *NoteHandler) DeleteNote(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteNote(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete note")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}
