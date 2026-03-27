package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"pf2e-companion/backend/services"
)

// PinGroupHandler holds the service dependency for pin group routes.
type PinGroupHandler struct {
	service services.PinGroupService
}

// NewPinGroupHandler constructs a PinGroupHandler.
func NewPinGroupHandler(service services.PinGroupService) *PinGroupHandler {
	return &PinGroupHandler{service: service}
}

// CreateGroup handles POST /games/:id/pin-groups.
func (h *PinGroupHandler) CreateGroup(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var body struct {
		PinIDs []uuid.UUID `json:"pin_ids"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if len(body.PinIDs) < 2 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "at least 2 pins required to create a group")
	}
	resp, err := h.service.CreateGroup(gameID, authUserID, body.PinIDs)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin not found")
		}
		return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
	}
	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListGameGroups handles GET /games/:id/pin-groups.
func (h *PinGroupHandler) ListGameGroups(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	gameID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	groups, err := h.service.ListGameGroups(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list pin groups")
	}
	return SuccessResponse(c, http.StatusOK, groups)
}

// UpdateGroup handles PATCH /pin-groups/:id.
func (h *PinGroupHandler) UpdateGroup(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	groupID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if colourVal, ok := updates["colour"]; ok {
		colour, ok := colourVal.(string)
		if !ok {
			return ErrorResponse(c, http.StatusBadRequest, "colour must be a string")
		}
		if err := ValidatePinColour(colour); err != nil {
			return ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
	}
	if iconVal, ok := updates["icon"]; ok {
		icon, ok := iconVal.(string)
		if !ok {
			return ErrorResponse(c, http.StatusBadRequest, "icon must be a string")
		}
		if err := ValidatePinIcon(icon); err != nil {
			return ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
	}
	resp, err := h.service.UpdateGroup(groupID, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin group not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update pin group")
	}
	return SuccessResponse(c, http.StatusOK, resp)
}

// AddPinToGroup handles POST /pin-groups/:id/pins.
func (h *PinGroupHandler) AddPinToGroup(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	groupID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	var body struct {
		PinID uuid.UUID `json:"pin_id"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.AddPinToGroup(groupID, body.PinID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin or group not found")
		}
		return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
	}
	return SuccessResponse(c, http.StatusOK, resp)
}

// RemovePinFromGroup handles DELETE /pin-groups/:id/pins/:pinId.
func (h *PinGroupHandler) RemovePinFromGroup(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	groupID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	pinID, err := ParseUUID(c, "pinId")
	if err != nil {
		return nil
	}
	resp, err := h.service.RemovePinFromGroup(groupID, pinID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin group not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to remove pin from group")
	}
	return SuccessResponse(c, http.StatusOK, resp)
}

// DisbandGroup handles DELETE /pin-groups/:id.
func (h *PinGroupHandler) DisbandGroup(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	groupID, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}
	if err := h.service.DisbandGroup(groupID, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "pin group not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to disband pin group")
	}
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "disbanded"})
}

// RegisterPinGroupRoutes wires all pin group routes on the given group.
func RegisterPinGroupRoutes(g *echo.Group, service services.PinGroupService) {
	h := NewPinGroupHandler(service)
	g.POST("/games/:id/pin-groups", h.CreateGroup)
	g.GET("/games/:id/pin-groups", h.ListGameGroups)
	g.PATCH("/pin-groups/:id", h.UpdateGroup)
	g.POST("/pin-groups/:id/pins", h.AddPinToGroup)
	g.DELETE("/pin-groups/:id/pins/:pinId", h.RemovePinFromGroup)
	g.DELETE("/pin-groups/:id", h.DisbandGroup)
}
