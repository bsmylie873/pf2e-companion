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

// CharacterHandler holds the service dependency for character-related route handlers.
type CharacterHandler struct {
	service services.CharacterService
}

// NewCharacterHandler constructs a CharacterHandler with the given service.
func NewCharacterHandler(service services.CharacterService) *CharacterHandler {
	return &CharacterHandler{service: service}
}

// RegisterCharacterRoutes wires all character routes onto the group.
func RegisterCharacterRoutes(g *echo.Group, service services.CharacterService) {
	h := NewCharacterHandler(service)
	g.POST("/games/:id/characters", h.CreateCharacter)
	g.GET("/games/:id/characters", h.ListGameCharacters)
	g.GET("/characters/:id", h.GetCharacter)
	g.PATCH("/characters/:id", h.UpdateCharacter)
	g.DELETE("/characters/:id", h.DeleteCharacter)
}

// CreateCharacter handles POST /games/:id/characters.
func (h *CharacterHandler) CreateCharacter(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var character models.Character
	if err := c.Bind(&character); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{
		"name": character.Name,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	resp, err := h.service.CreateCharacter(gameID, authUserID, &character)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create character")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameCharacters handles GET /games/:id/characters.
func (h *CharacterHandler) ListGameCharacters(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	characters, err := h.service.ListGameCharacters(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list characters")
	}

	return SuccessResponse(c, http.StatusOK, characters)
}

// GetCharacter handles GET /characters/:id.
func (h *CharacterHandler) GetCharacter(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	character, err := h.service.GetCharacter(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "character not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to get character")
	}

	return SuccessResponse(c, http.StatusOK, character)
}

// UpdateCharacter handles PATCH /characters/:id.
func (h *CharacterHandler) UpdateCharacter(c echo.Context) error {
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

	character, err := h.service.UpdateCharacter(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "character not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update character")
	}

	return SuccessResponse(c, http.StatusOK, character)
}

// DeleteCharacter handles DELETE /characters/:id.
func (h *CharacterHandler) DeleteCharacter(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteCharacter(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "character not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete character")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}
