package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/services"
)

// FolderHandler handles HTTP requests for folders.
type FolderHandler struct {
	service services.FolderService
}

// NewFolderHandler creates a new FolderHandler.
func NewFolderHandler(service services.FolderService) *FolderHandler {
	return &FolderHandler{service: service}
}

// RegisterFolderRoutes wires all folder routes onto the group.
func RegisterFolderRoutes(g *echo.Group, service services.FolderService) {
	h := NewFolderHandler(service)
	g.POST("/games/:id/folders", h.CreateFolder)
	g.GET("/games/:id/folders", h.ListFolders)
	g.PATCH("/folders/:id", h.RenameFolder)
	g.DELETE("/folders/:id", h.DeleteFolder)
	g.PUT("/games/:id/folders/reorder", h.ReorderFolders)
}

// CreateFolder creates a new folder within a game.
// POST /games/:id/folders
func (h *FolderHandler) CreateFolder(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var body struct {
		Name       string `json:"name"`
		FolderType string `json:"folder_type"`
		Visibility string `json:"visibility"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{"name": body.Name, "folder_type": body.FolderType})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	if body.Visibility == "" {
		body.Visibility = "game-wide"
	}

	folder, err := h.service.CreateFolder(gameID, authUserID, body.Name, body.FolderType, body.Visibility)
	if err != nil {
		return mapFolderError(c, err)
	}

	return SuccessResponse(c, http.StatusCreated, folder)
}

// ListFolders returns folders for a game, filtered by type query param.
// GET /games/:id/folders?type=session|note
func (h *FolderHandler) ListFolders(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	folderType := c.QueryParam("type")
	if folderType != "session" && folderType != "note" {
		return ErrorResponse(c, http.StatusBadRequest, "query param 'type' must be 'session' or 'note'")
	}

	folders, err := h.service.ListFolders(gameID, authUserID, folderType)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch folders")
	}

	return SuccessResponse(c, http.StatusOK, folders)
}

// RenameFolder renames a folder.
// PATCH /folders/:id
func (h *FolderHandler) RenameFolder(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{"name": body.Name})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	folder, err := h.service.RenameFolder(id, authUserID, body.Name)
	if err != nil {
		return mapFolderError(c, err)
	}

	return SuccessResponse(c, http.StatusOK, folder)
}

// DeleteFolder removes a folder (contents are unassigned, not deleted).
// DELETE /folders/:id
func (h *FolderHandler) DeleteFolder(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteFolder(id, authUserID); err != nil {
		return mapFolderError(c, err)
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

// ReorderFolders batch-updates folder positions.
// PUT /games/:id/folders/reorder
func (h *FolderHandler) ReorderFolders(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var body struct {
		FolderType string   `json:"folder_type"`
		FolderIDs  []string `json:"folder_ids"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if body.FolderType != "session" && body.FolderType != "note" {
		return ErrorResponse(c, http.StatusBadRequest, "folder_type must be 'session' or 'note'")
	}

	if len(body.FolderIDs) == 0 {
		return ErrorResponse(c, http.StatusBadRequest, "folder_ids must not be empty")
	}

	orderedIDs := make([]uuid.UUID, 0, len(body.FolderIDs))
	for _, raw := range body.FolderIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid UUID in folder_ids: "+raw)
		}
		orderedIDs = append(orderedIDs, id)
	}

	if err := h.service.ReorderFolders(gameID, authUserID, body.FolderType, orderedIDs); err != nil {
		return mapFolderError(c, err)
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "reordered"})
}

func mapFolderError(c echo.Context, err error) error {
	if errors.Is(err, services.ErrForbidden) {
		return ErrorResponse(c, http.StatusForbidden, "forbidden")
	}
	if errors.Is(err, services.ErrSessionFoldersReadOnly) {
		return ErrorResponse(c, http.StatusForbidden, "session_folders_read_only")
	}
	if errors.Is(err, services.ErrConflict) {
		return ErrorResponse(c, http.StatusConflict, "a folder with this name already exists")
	}
	if errors.Is(err, services.ErrValidation) {
		return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, http.StatusNotFound, "folder not found")
	}
	return ErrorResponse(c, http.StatusInternalServerError, "failed to process folder request")
}
