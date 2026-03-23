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

// RegisterNoteRoutes wires all note routes onto the group.
func RegisterNoteRoutes(g *echo.Group, service services.NoteService) {
	h := NewNoteHandler(service)
	g.POST("/games/:id/notes", h.CreateGameNote)
	g.POST("/users/:id/notes", h.CreateUserNote)
	g.GET("/games/:id/notes", h.ListGameNotes)
	g.GET("/users/:id/notes", h.ListUserNotes)
	g.GET("/notes/:id", h.GetNote)
	g.PATCH("/notes/:id", h.UpdateNote)
	g.DELETE("/notes/:id", h.DeleteNote)
}

// CreateGameNote creates a shared note belonging to a game.
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

	resp, err := h.service.CreateGameNote(gameID, authUserID, &note)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create note")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// CreateUserNote creates a private note belonging to a user.
// POST /users/:id/notes
func (h *NoteHandler) CreateUserNote(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	pathUserID, err := ParseUUID(c, "id")
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

	resp, err := h.service.CreateUserNote(pathUserID, authUserID, &note)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create note")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameNotes returns all notes belonging to a game.
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

	notes, err := h.service.ListGameNotes(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch notes")
	}

	return SuccessResponse(c, http.StatusOK, notes)
}

// ListUserNotes returns all notes belonging to a user.
// GET /users/:id/notes
func (h *NoteHandler) ListUserNotes(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	pathUserID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	notes, err := h.service.ListUserNotes(pathUserID, authUserID)
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
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
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

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}
