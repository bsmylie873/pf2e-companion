package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	authmw "pf2e-companion/backend/middleware"
	"pf2e-companion/backend/models"
	"pf2e-companion/backend/services"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	service services.AuthService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(service services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	missing := ValidateRequired(map[string]interface{}{
		"username": req.Username,
		"email":    req.Email,
		"password": req.Password,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}
	user, pair, err := h.service.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ErrorResponse(c, http.StatusUnprocessableEntity, "username or email already taken")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to register")
	}
	secure := authmw.CookieSecure()
	authmw.SetAccessCookie(c, pair.AccessToken, secure)
	authmw.SetRefreshCookie(c, pair.RefreshToken, secure)
	return SuccessResponse(c, http.StatusCreated, user)
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	missing := ValidateRequired(map[string]interface{}{
		"username": req.Username,
		"password": req.Password,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}
	user, pair, err := h.service.Login(req)
	if err != nil {
		if err.Error() == "invalid credentials" {
			return ErrorResponse(c, http.StatusUnauthorized, "invalid credentials")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "login failed")
	}
	secure := authmw.CookieSecure()
	authmw.SetAccessCookie(c, pair.AccessToken, secure)
	authmw.SetRefreshCookie(c, pair.RefreshToken, secure)
	return SuccessResponse(c, http.StatusOK, user)
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		_ = h.service.Logout(cookie.Value)
	}
	authmw.ClearAuthCookies(c)
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "logged out"})
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		return ErrorResponse(c, http.StatusUnauthorized, "no refresh token")
	}
	pair, err := h.service.RefreshTokens(cookie.Value)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "invalid or expired refresh token")
	}
	secure := authmw.CookieSecure()
	authmw.SetAccessCookie(c, pair.AccessToken, secure)
	authmw.SetRefreshCookie(c, pair.RefreshToken, secure)
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "refreshed"})
}

// Me handles GET /auth/me.
func (h *AuthHandler) Me(c echo.Context) error {
	userID, err := GetAuthUserID(c)
	if err != nil {
		return nil
	}
	user, err := h.service.GetMe(userID)
	if err != nil {
		return ErrorResponse(c, http.StatusNotFound, "user not found")
	}
	return SuccessResponse(c, http.StatusOK, user)
}

// ForgotPassword handles POST /auth/forgot-password.
// Always returns 200 to prevent user enumeration.
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return SuccessResponse(c, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})
	}
	// Intentionally ignore error to prevent user enumeration
	_ = h.service.RequestPasswordReset(req.Email)
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "if that email exists, a reset link has been sent"})
}

// ResetPassword handles POST /auth/reset-password.
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req models.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	missing := ValidateRequired(map[string]interface{}{
		"token":        req.Token,
		"new_password": req.NewPassword,
	})
	if len(missing) > 0 {
		return ErrorResponse(c, http.StatusUnprocessableEntity, "missing required fields: "+strings.Join(missing, ", "))
	}
	if err := h.service.ResetPassword(req.Token, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "already used") {
			return ErrorResponse(c, http.StatusUnprocessableEntity, err.Error())
		}
		return ErrorResponse(c, http.StatusInternalServerError, "failed to reset password")
	}
	return SuccessResponse(c, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

// RegisterAuthRoutes wires auth routes onto the Echo instance and protected group.
// loginRateLimiter is applied to login; passwordResetRateLimiter is applied to forgot-password.
func RegisterAuthRoutes(e *echo.Echo, g *echo.Group, service services.AuthService, loginRateLimiter echo.MiddlewareFunc, passwordResetRateLimiter echo.MiddlewareFunc) {
	h := NewAuthHandler(service)
	e.POST("/auth/register", h.Register)
	e.POST("/auth/login", h.Login, loginRateLimiter)
	e.POST("/auth/refresh", h.Refresh)
	e.POST("/auth/forgot-password", h.ForgotPassword, passwordResetRateLimiter)
	e.POST("/auth/reset-password", h.ResetPassword)
	g.POST("/auth/logout", h.Logout)
	g.GET("/auth/me", h.Me)
}
