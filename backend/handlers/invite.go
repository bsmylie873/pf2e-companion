package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"pf2e-companion/backend/services"
)

// InviteHandler handles invite-related HTTP requests.
type InviteHandler struct {
	service services.InviteService
}

// NewInviteHandler constructs an InviteHandler.
func NewInviteHandler(service services.InviteService) *InviteHandler {
	return &InviteHandler{service: service}
}

// generateInviteRequest is the request body for POST /games/:gameId/invite.
type generateInviteRequest struct {
	ExpiresIn string `json:"expires_in"`
}

// GenerateInvite handles POST /games/:gameId/invite.
func (h *InviteHandler) GenerateInvite(c echo.Context) error {
	gameID, err := ParseUUID(c, "gameId")
	if err != nil {
		return nil
	}
	callerID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	var req generateInviteRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if req.ExpiresIn == "" {
		req.ExpiresIn = "never"
	}
	resp, err := h.service.GenerateInvite(gameID, callerID, req.ExpiresIn)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, services.ErrValidation) {
			return ErrorResponse(c, http.StatusUnprocessableEntity, "invalid expires_in value; use \"24h\", \"7d\", or \"never\"")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to generate invite")
	}
	return SuccessResponse(c, http.StatusCreated, resp)
}

// GetActiveInvite handles GET /games/:gameId/invite.
func (h *InviteHandler) GetActiveInvite(c echo.Context) error {
	gameID, err := ParseUUID(c, "gameId")
	if err != nil {
		return nil
	}
	callerID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	resp, err := h.service.GetActiveInvite(gameID, callerID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to get invite status")
	}
	return SuccessResponse(c, http.StatusOK, resp)
}

// RevokeInvite handles DELETE /games/:gameId/invite.
func (h *InviteHandler) RevokeInvite(c echo.Context) error {
	gameID, err := ParseUUID(c, "gameId")
	if err != nil {
		return nil
	}
	callerID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	if err := h.service.RevokeInvite(gameID, callerID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to revoke invite")
	}
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "invite revoked"})
}

// ValidateInvite handles GET /invite/:token (public).
func ValidateInvite(service services.InviteService) echo.HandlerFunc {
	return func(c echo.Context) error {
		rawToken := c.Param("token")
		if rawToken == "" {
			return ErrorResponse(c, http.StatusBadRequest, "missing token")
		}
		resp, err := service.ValidateInvite(rawToken)
		if err != nil {
			if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "revoked") || strings.Contains(err.Error(), "expired") {
				return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
			}
			return ErrorResponse(c, http.StatusInternalServerError, "failed to validate invite")
		}
		return SuccessResponse(c, http.StatusOK, resp)
	}
}

// RedeemInvite handles POST /invite/:token/redeem (protected).
func RedeemInvite(service services.InviteService) echo.HandlerFunc {
	return func(c echo.Context) error {
		rawToken := c.Param("token")
		if rawToken == "" {
			return ErrorResponse(c, http.StatusBadRequest, "missing token")
		}
		userID, err := GetAuthUserID(c)
		if err != nil {
			return nil
		}
		resp, err := service.RedeemInvite(rawToken, userID)
		if err != nil {
			if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "revoked") || strings.Contains(err.Error(), "expired") {
				return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
			}
			return ErrorResponse(c, http.StatusInternalServerError, "failed to redeem invite")
		}
		return SuccessResponse(c, http.StatusOK, resp)
	}
}

// RegisterInviteRoutes wires GM-only invite management routes onto the protected group.
func RegisterInviteRoutes(g *echo.Group, service services.InviteService) {
	h := NewInviteHandler(service)
	g.POST("/games/:gameId/invite", h.GenerateInvite)
	g.GET("/games/:gameId/invite", h.GetActiveInvite)
	g.DELETE("/games/:gameId/invite", h.RevokeInvite)
}
