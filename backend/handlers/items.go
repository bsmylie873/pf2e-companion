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

// ItemHandler holds the service dependency for item-related route handlers.
type ItemHandler struct {
	service services.ItemService
}

// NewItemHandler constructs an ItemHandler with the given service.
func NewItemHandler(service services.ItemService) *ItemHandler {
	return &ItemHandler{service: service}
}

// RegisterItemRoutes registers all item routes on the group.
func RegisterItemRoutes(g *echo.Group, service services.ItemService) {
	h := NewItemHandler(service)
	g.POST("/games/:id/items", h.CreateItem)
	g.GET("/games/:id/items", h.ListGameItems)
	g.GET("/characters/:id/items", h.ListCharacterItems)
	g.GET("/items/:id", h.GetItem)
	g.PATCH("/items/:id", h.UpdateItem)
	g.DELETE("/items/:id", h.DeleteItem)
}

// CreateItem handles POST /games/:id/items.
func (h *ItemHandler) CreateItem(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var item models.Item
	if err := c.Bind(&item); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{"name": item.Name})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	resp, err := h.service.CreateItem(gameID, authUserID, &item)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create item")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameItems handles GET /games/:id/items.
func (h *ItemHandler) ListGameItems(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	items, err := h.service.ListGameItems(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch items")
	}

	return SuccessResponse(c, http.StatusOK, items)
}

// ListCharacterItems handles GET /characters/:id/items.
func (h *ItemHandler) ListCharacterItems(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	characterID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	items, err := h.service.ListCharacterItems(characterID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch items")
	}

	return SuccessResponse(c, http.StatusOK, items)
}

// GetItem handles GET /items/:id.
func (h *ItemHandler) GetItem(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	item, err := h.service.GetItem(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "item not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch item")
	}

	return SuccessResponse(c, http.StatusOK, item)
}

// UpdateItem handles PATCH /items/:id.
func (h *ItemHandler) UpdateItem(c echo.Context) error {
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

	item, err := h.service.UpdateItem(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "item not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update item")
	}

	return SuccessResponse(c, http.StatusOK, item)
}

// DeleteItem handles DELETE /items/:id.
func (h *ItemHandler) DeleteItem(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteItem(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "item not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete item")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}
