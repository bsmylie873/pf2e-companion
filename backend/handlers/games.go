package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

type createGameMember struct {
	UserID string `json:"user_id"`
	IsGM   bool   `json:"is_gm"`
}

type createGameRequest struct {
	Title          string             `json:"title"`
	Description    *string            `json:"description"`
	SplashImageURL *string            `json:"splash_image_url"`
	Members        []createGameMember `json:"members"`
}

type GameHandler struct {
	service services.GameService
}

func NewGameHandler(service services.GameService) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) CreateGame(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	var req createGameRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{
		"title": req.Title,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	game := models.Game{
		Title:          req.Title,
		Description:    req.Description,
		SplashImageURL: req.SplashImageURL,
	}

	var memberships []models.GameMembership
	for _, m := range req.Members {
		uid, err := uuid.Parse(m.UserID)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "invalid user_id: "+m.UserID)
		}
		memberships = append(memberships, models.GameMembership{
			UserID: uid,
			IsGM:   m.IsGM,
		})
	}

	resp, err := h.service.CreateGame(&game, memberships, authUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create game")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

func (h *GameHandler) ListGames(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	games, err := h.service.ListGames(authUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list games")
	}
	return SuccessResponse(c, http.StatusOK, games)
}

func (h *GameHandler) GetGame(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	game, err := h.service.GetGame(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to get game")
	}

	return SuccessResponse(c, http.StatusOK, game)
}

func (h *GameHandler) UpdateGame(c echo.Context) error {
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

	game, err := h.service.UpdateGame(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update game")
	}

	return SuccessResponse(c, http.StatusOK, game)
}

func (h *GameHandler) DeleteGame(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteGame(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "game not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete game")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

func RegisterGameRoutes(g *echo.Group, service services.GameService) {
	h := NewGameHandler(service)
	g.POST("/games", h.CreateGame)
	g.GET("/games", h.ListGames)
	g.GET("/games/:id", h.GetGame)
	g.PATCH("/games/:id", h.UpdateGame)
	g.DELETE("/games/:id", h.DeleteGame)
}
