package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/repositories"
)

const maxMapImageSize = 10 * 1024 * 1024 // 10 MB

type MapImageHandler struct {
	gameRepo       repositories.GameRepository
	membershipRepo repositories.MembershipRepository
}

func NewMapImageHandler(gameRepo repositories.GameRepository, membershipRepo repositories.MembershipRepository) *MapImageHandler {
	return &MapImageHandler{gameRepo: gameRepo, membershipRepo: membershipRepo}
}

// UploadMapImage handles POST /games/:id/map-image.
// Only the GM of the game may upload or replace the map image.
func (h *MapImageHandler) UploadMapImage(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	// Check GM membership
	membership, err := h.membershipRepo.FindByUserAndGameID(authUserID, gameID)
	if err != nil {
		return ErrorResponse(c, http.StatusForbidden, "forbidden")
	}
	if !membership.IsGM {
		return ErrorResponse(c, http.StatusForbidden, "forbidden")
	}

	game, err := h.gameRepo.FindByID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch game")
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

	if fileHeader.Size > maxMapImageSize {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "file too large: maximum 10MB")
	}

	// Delete old file if present
	if game.MapImageURL != nil {
		_ = os.Remove("." + *game.MapImageURL)
	}

	filename := fmt.Sprintf("%s-%s.%s", gameID, uuid.New().String()[:8], ext)
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

	mapURL := "/uploads/maps/" + filename
	updatedGame, err := h.gameRepo.Update(gameID, map[string]interface{}{"map_image_url": mapURL})
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update game")
	}

	return SuccessResponse(c, http.StatusOK, updatedGame)
}

// DeleteMapImage handles DELETE /games/:id/map-image.
// Only the GM of the game may delete the map image.
func (h *MapImageHandler) DeleteMapImage(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	membership, err := h.membershipRepo.FindByUserAndGameID(authUserID, gameID)
	if err != nil {
		return ErrorResponse(c, http.StatusForbidden, "forbidden")
	}
	if !membership.IsGM {
		return ErrorResponse(c, http.StatusForbidden, "forbidden")
	}

	game, err := h.gameRepo.FindByID(gameID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch game")
	}

	if game.MapImageURL != nil {
		_ = os.Remove("." + *game.MapImageURL)
	}

	if _, err := h.gameRepo.Update(gameID, map[string]interface{}{"map_image_url": nil}); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update game")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

func RegisterMapImageRoutes(g *echo.Group, gameRepo repositories.GameRepository, membershipRepo repositories.MembershipRepository) {
	h := NewMapImageHandler(gameRepo, membershipRepo)
	g.POST("/games/:id/map-image", h.UploadMapImage)
	g.DELETE("/games/:id/map-image", h.DeleteMapImage)
}
