package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

// BackupHandler holds the service dependency for backup-related routes.
type BackupHandler struct {
	service services.BackupService
}

// NewBackupHandler constructs a BackupHandler with the given service.
func NewBackupHandler(service services.BackupService) *BackupHandler {
	return &BackupHandler{service: service}
}

// ExportGame handles GET /games/:id/backup/export.
// Returns a JSON file download containing all visible sessions and notes.
func (h *BackupHandler) ExportGame(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	backup, err := h.service.ExportGame(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to export game")
	}

	filename := fmt.Sprintf("game-%s-backup-%s.json", gameID, time.Now().UTC().Format("2006-01-02"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.JSON(http.StatusOK, backup)
}

// ExportSession handles GET /sessions/:id/backup/export.
// Returns a JSON file download containing a single session and its associated notes.
func (h *BackupHandler) ExportSession(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	sessionID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	backup, err := h.service.ExportSession(sessionID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "session not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to export session")
	}

	filename := fmt.Sprintf("session-%s-backup-%s.json", sessionID, time.Now().UTC().Format("2006-01-02"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.JSON(http.StatusOK, backup)
}

// ExportNote handles GET /notes/:id/backup/export.
// Returns a JSON file download containing a single note.
func (h *BackupHandler) ExportNote(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	noteID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	backup, err := h.service.ExportNote(noteID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "note not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to export note")
	}

	filename := fmt.Sprintf("note-%s-backup-%s.json", noteID, time.Now().UTC().Format("2006-01-02"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	return c.JSON(http.StatusOK, backup)
}

// ImportGame handles POST /games/:id/backup/import.
// Accepts a BackupFile JSON body and imports records into the target game.
func (h *BackupHandler) ImportGame(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	// Validate mode query parameter.
	mode := c.QueryParam("mode")
	if mode != "merge" && mode != "overwrite" {
		return ErrorResponse(c, http.StatusBadRequest, "mode must be merge or overwrite")
	}

	// Enforce 10 MB body limit.
	const maxBodySize = 10 << 20 // 10 MB
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxBodySize)

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "file exceeds 10MB or could not be read")
	}

	var backup models.BackupFile
	if err := json.Unmarshal(body, &backup); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid JSON")
	}

	// Validate schema version.
	if backup.SchemaVersion != "1" {
		return ErrorResponse(c, http.StatusBadRequest, "unsupported schema version")
	}

	// Validate game_id matches URL param.
	if backup.GameID != gameID {
		return ErrorResponse(c, http.StatusBadRequest, "game_id mismatch")
	}

	summary, err := h.service.ImportGame(gameID, authUserID, mode, &backup)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to import backup")
	}

	return SuccessResponse(c, http.StatusCreated, summary)
}

// RegisterBackupRoutes registers all backup-related routes on the group.
func RegisterBackupRoutes(g *echo.Group, service services.BackupService, backupRateLimiter echo.MiddlewareFunc) {
	h := NewBackupHandler(service)
	g.GET("/games/:id/backup/export", h.ExportGame, backupRateLimiter)
	g.GET("/sessions/:id/backup/export", h.ExportSession, backupRateLimiter)
	g.GET("/notes/:id/backup/export", h.ExportNote, backupRateLimiter)
	g.POST("/games/:id/backup/import", h.ImportGame, backupRateLimiter)
}
