package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	custmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	_, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	responses, err := h.service.ListUsers()
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve users")
	}
	return SuccessResponse(c, http.StatusOK, responses)
}

func (h *UserHandler) GetUser(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	resp, err := h.service.GetUser(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "user not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve user")
	}

	if id == authUserID {
		return SuccessResponse(c, http.StatusOK, resp)
	}

	public := models.UserPublicResponse{ID: resp.ID, Username: resp.Username, AvatarURL: resp.AvatarURL}
	return SuccessResponse(c, http.StatusOK, public)
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if id != authUserID {
		return ErrorResponse(c, http.StatusForbidden, "forbidden")
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	_, hasPassword := updates["password"]

	resp, err := h.service.UpdateUser(id, updates, authUserID)
	if err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "user not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update user")
	}

	if hasPassword {
		custmw.ClearAuthCookies(c)
	}

	return SuccessResponse(c, http.StatusOK, resp)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	authUserID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}

	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteUser(id, authUserID); err != nil {
		if errors.Is(err, services.ErrForbidden) {
			return ErrorResponse(c, http.StatusForbidden, "forbidden")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "user not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete user")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

func RegisterUserRoutes(g *echo.Group, service services.UserService) {
	h := NewUserHandler(service)
	g.GET("/users", h.ListUsers)
	g.GET("/users/:id", h.GetUser)
	g.PATCH("/users/:id", h.UpdateUser)
	g.DELETE("/users/:id", h.DeleteUser)
}
