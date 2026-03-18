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

type GameHandler struct {
	service services.GameService
}

func NewGameHandler(service services.GameService) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) CreateGame(c echo.Context) error {
	var game models.Game
	if err := c.Bind(&game); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{
		"title": game.Title,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	resp, err := h.service.CreateGame(&game)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create game")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

func (h *GameHandler) ListGames(c echo.Context) error {
	games, err := h.service.ListGames()
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list games")
	}
	return SuccessResponse(c, http.StatusOK, games)
}

func (h *GameHandler) GetGame(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	game, err := h.service.GetGame(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to get game")
	}

	return SuccessResponse(c, http.StatusOK, game)
}

func (h *GameHandler) UpdateGame(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	game, err := h.service.UpdateGame(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update game")
	}

	return SuccessResponse(c, http.StatusOK, game)
}

func (h *GameHandler) DeleteGame(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteGame(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete game")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

func RegisterGameRoutes(e *echo.Echo, service services.GameService) {
	h := NewGameHandler(service)
	e.POST("/games", h.CreateGame)
	e.GET("/games", h.ListGames)
	e.GET("/games/:id", h.GetGame)
	e.PATCH("/games/:id", h.UpdateGame)
	e.DELETE("/games/:id", h.DeleteGame)
}
