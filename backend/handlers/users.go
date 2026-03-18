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

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	missing := ValidateRequired(map[string]interface{}{
		"username":      user.Username,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}

	resp, err := h.service.CreateUser(&user)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to create user")
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	responses, err := h.service.ListUsers()
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve users")
	}
	return SuccessResponse(c, http.StatusOK, responses)
}

func (h *UserHandler) GetUser(c echo.Context) error {
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

	return SuccessResponse(c, http.StatusOK, resp)
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	var updates map[string]interface{}
	if err := c.Bind(&updates); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	resp, err := h.service.UpdateUser(id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "user not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update user")
	}

	return SuccessResponse(c, http.StatusOK, resp)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	id, err := ParseUUID(c, "id")
	if err != nil {
		return nil
	}

	if err := h.service.DeleteUser(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "user not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to delete user")
	}

	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "deleted"})
}

func RegisterUserRoutes(e *echo.Echo, service services.UserService) {
	h := NewUserHandler(service)
	e.POST("/users", h.CreateUser)
	e.GET("/users", h.ListUsers)
	e.GET("/users/:id", h.GetUser)
	e.PATCH("/users/:id", h.UpdateUser)
	e.DELETE("/users/:id", h.DeleteUser)
}
