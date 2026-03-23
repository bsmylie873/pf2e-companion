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

// createMembershipRequest is used to bind JSON for CreateMembership,
// keeping GameID and UserID as strings so we can parse them manually.
type createMembershipRequest struct {
	GameID string `json:"game_id"`
	UserID string `json:"user_id"`
	IsGM   bool   `json:"is_gm"`
}

// MembershipHandler holds the service dependency for membership routes.
type MembershipHandler struct {
	service services.MembershipService
}

// NewMembershipHandler constructs a MembershipHandler with the given service.
func NewMembershipHandler(service services.MembershipService) *MembershipHandler {
	return &MembershipHandler{service: service}
}

// CreateMembership handles POST /memberships.
func (h *MembershipHandler) CreateMembership(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	var req createMembershipRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{
		"game_id": req.GameID,
		"user_id": req.UserID,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	gameID, err := uuid.Parse(req.GameID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid UUID for game_id: "+req.GameID)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid UUID for user_id: "+req.UserID)
	}

	membership := models.GameMembership{
		GameID: gameID,
		UserID: userID,
		IsGM:   req.IsGM,
	}

	resp, err := h.service.CreateMembership(&membership, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create membership")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// ListMemberships handles GET /memberships?game_id=<uuid>.
func (h *MembershipHandler) ListMemberships(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	gameIDStr := c.QueryParam("game_id")
	if gameIDStr == "" {
		return ErrorResponse(c, http.StatusBadRequest, "query param game_id is required")
	}

	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid UUID for game_id: "+gameIDStr)
	}

	memberships, err := h.service.ListMemberships(gameID, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to list memberships")
	}

	return SuccessResponse(c, http.StatusOK, memberships)
}

// GetMembership handles GET /memberships/:id.
func (h *MembershipHandler) GetMembership(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	membership, err := h.service.GetMembership(id, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "membership not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to fetch membership")
	}

	return SuccessResponse(c, http.StatusOK, membership)
}

// UpdateMembership handles PATCH /memberships/:id.
func (h *MembershipHandler) UpdateMembership(c echo.Context) error {
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

	membership, err := h.service.UpdateMembership(id, authUserID, updates)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "membership not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update membership")
	}

	return SuccessResponse(c, http.StatusOK, membership)
}

// DeleteMembership handles DELETE /memberships/:id.
func (h *MembershipHandler) DeleteMembership(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteMembership(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "membership not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete membership")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

// RegisterMembershipRoutes wires up all membership routes on the group.
func RegisterMembershipRoutes(g *echo.Group, service services.MembershipService) {
	h := NewMembershipHandler(service)
	g.POST("/memberships", h.CreateMembership)
	g.GET("/memberships", h.ListMemberships)
	g.GET("/memberships/:id", h.GetMembership)
	g.PATCH("/memberships/:id", h.UpdateMembership)
	g.DELETE("/memberships/:id", h.DeleteMembership)
}
