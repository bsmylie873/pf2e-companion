package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	authpkg "pf2e-companion/backend/auth"
	"pf2e-companion/backend/services"
)

const maxMapFileSize = 10 * 1024 * 1024 // 10 MB

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// MapHandler handles map-related routes.
type MapHandler struct {
	service services.MapService
	hub     *MapEventHub
}

// NewMapHandler constructs a MapHandler.
func NewMapHandler(service services.MapService, hub *MapEventHub) *MapHandler {
	return &MapHandler{service: service, hub: hub}
}

// ListMaps handles GET /games/:id/maps.
func (h *MapHandler) ListMaps(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	maps, err := h.service.ListMaps(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list maps")
	}
	return SuccessResponse(c, http.StatusOK, maps)
}

// ListArchivedMaps handles GET /games/:id/maps/archived.
func (h *MapHandler) ListArchivedMaps(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	maps, err := h.service.ListArchivedMaps(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list archived maps")
	}
	return SuccessResponse(c, http.StatusOK, maps)
}

// CreateMap handles POST /games/:id/maps.
func (h *MapHandler) CreateMap(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var body struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if body.Name == "" {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "name is required")
	}
	m, err := h.service.CreateMap(gameID, authUserID, body.Name, body.Description)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, services.ErrConflict) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, "a map with that name already exists in this campaign")
		}
		if errors.Is(err, services.ErrValidation) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, "name is required")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create map")
	}
	h.hub.Broadcast(gameID, MapEvent{Type: "map_created", MapID: m.ID, GameID: gameID, Data: m})
	return SuccessResponse(c, http.StatusCreated, m)
}

// RenameMap handles PATCH /games/:id/maps/:mapId.
func (h *MapHandler) RenameMap(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	mapID, err := ParseUUID(c, "mapId")
	if err != nil {
		return nil
	}
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	name := ""
	if body.Name != nil {
		name = *body.Name
	}
	m, err := h.service.RenameMap(mapID, authUserID, name, body.Description)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, services.ErrConflict) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, "a map with that name already exists in this campaign")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "map not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update map")
	}
	h.hub.Broadcast(gameID, MapEvent{Type: "map_renamed", MapID: mapID, GameID: gameID, Data: m})
	return SuccessResponse(c, http.StatusOK, m)
}

// ReorderMaps handles PATCH /games/:id/maps/order.
func (h *MapHandler) ReorderMaps(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var body struct {
		MapIDs []uuid.UUID `json:"map_ids"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if len(body.MapIDs) == 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "map_ids is required")
	}
	if err := h.service.ReorderMaps(gameID, authUserID, body.MapIDs); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to reorder maps")
	}
	h.hub.Broadcast(gameID, MapEvent{Type: "map_reordered", GameID: gameID, Data: body.MapIDs})
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "reordered"})
}

// ArchiveMap handles DELETE /games/:id/maps/:mapId.
func (h *MapHandler) ArchiveMap(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	mapID, err := ParseUUID(c, "mapId")
	if err != nil {
		return nil
	}
	if err := h.service.ArchiveMap(mapID, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "map not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to archive map")
	}
	h.hub.Broadcast(gameID, MapEvent{Type: "map_archived", MapID: mapID, GameID: gameID, Data: nil})
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "archived"})
}

// RestoreMap handles POST /games/:id/maps/:mapId/restore.
func (h *MapHandler) RestoreMap(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	mapID, err := ParseUUID(c, "mapId")
	if err != nil {
		return nil
	}
	m, err := h.service.RestoreMap(mapID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "map not found or no longer restorable")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to restore map")
	}
	h.hub.Broadcast(gameID, MapEvent{Type: "map_restored", MapID: mapID, GameID: gameID, Data: m})
	return SuccessResponse(c, http.StatusOK, m)
}

// UploadMapImage handles POST /games/:id/maps/:mapId/image.
func (h *MapHandler) UploadMapImage(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	_, err = ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	mapID, err := ParseUUID(c, "mapId")
	if err != nil {
		return nil
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "missing file")
	}
	contentType := fileHeader.Header.Get("Content-Type")
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/webp":
		ext = "webp"
	default:
		return ErrorResponse(c, http.StatusUnprocessableEntity, "unsupported image type: must be jpeg, png, or webp")
	}
	if fileHeader.Size > maxMapFileSize {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "file too large: maximum 10MB")
	}
	filename := fmt.Sprintf("%s-%s.%s", mapID, uuid.New().String()[:8], ext)
	destPath := "./uploads/maps/" + filename
	if err := os.MkdirAll("./uploads/maps", 0755); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create upload directory")
	}
	src, err := fileHeader.Open()
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to open upload")
	}
	defer src.Close()
	dst, err := os.Create(destPath)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create destination file")
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to save file")
	}
	imageURL := "/uploads/maps/" + filename
	m, err := h.service.SetMapImage(mapID, authUserID, imageURL)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "map not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update map image")
	}
	h.hub.Broadcast(m.GameID, MapEvent{Type: "map_renamed", MapID: mapID, GameID: m.GameID, Data: m})
	return SuccessResponse(c, http.StatusOK, m)
}

// MapWebSocket handles GET /games/:id/maps/ws.
// Authenticates via access_token cookie (inline JWT check, since WS upgrade needs special handling).
func (h *MapHandler) MapWebSocket(c echo.Context) error {
	cookie, err := c.Cookie("access_token")
	if err != nil || cookie.Value == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized"})
	}
	claims, err := authpkg.ValidateToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": 401, "message": "unauthorized"})
	}
	_ = claims

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	ws, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	h.hub.Register(gameID, ws)
	defer h.hub.Unregister(gameID, ws)

	// Keep connection open, reading (and discarding) client messages for ping/pong
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
	return nil
}

// RegisterMapRoutes wires all map-related routes.
func RegisterMapRoutes(e *echo.Echo, g *echo.Group, service services.MapService, hub *MapEventHub) {
	h := NewMapHandler(service, hub)
	g.GET("/games/:id/maps", h.ListMaps)
	g.GET("/games/:id/maps/archived", h.ListArchivedMaps)
	g.POST("/games/:id/maps", h.CreateMap)
	g.PATCH("/games/:id/maps/:mapId", h.RenameMap)
	g.PATCH("/games/:id/maps/order", h.ReorderMaps)
	g.DELETE("/games/:id/maps/:mapId", h.ArchiveMap)
	g.POST("/games/:id/maps/:mapId/restore", h.RestoreMap)
	g.POST("/games/:id/maps/:mapId/image", h.UploadMapImage)
	// WS endpoint registered on base echo (not protected group — auth is handled inline)
	e.GET("/games/:id/maps/ws", h.MapWebSocket)
}
