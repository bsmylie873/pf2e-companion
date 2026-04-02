package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	authpkg "pf2e-companion/backend/auth"
	"pf2e-companion/backend/services"
)

// AuthUserIDKey is the Echo context key for the authenticated user's UUID.
const AuthUserIDKey = "auth_user_id"

// CookieSecure returns true if COOKIE_SECURE env var is "true" or "1".
func CookieSecure() bool {
	v := strings.ToLower(os.Getenv("COOKIE_SECURE"))
	return v == "true" || v == "1"
}

// SetAccessCookie writes the access_token cookie (HttpOnly, 15min).
func SetAccessCookie(c echo.Context, value string, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    value,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
		MaxAge:   900,
	})
}

// SetRefreshCookie writes the refresh_token cookie scoped to /auth/refresh (7 days).
func SetRefreshCookie(c echo.Context, value string, secure bool) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		Path:     "/auth/refresh",
		MaxAge:   604800,
	})
}

// ClearAuthCookies clears all auth-related cookies.
func ClearAuthCookies(c echo.Context) {
	secure := CookieSecure()
	c.SetCookie(&http.Cookie{Name: "access_token", Value: "", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteNoneMode, Path: "/"})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, HttpOnly: true, Secure: secure, SameSite: http.SameSiteNoneMode, Path: "/auth/refresh"})
	c.SetCookie(&http.Cookie{Name: "csrf_token", Value: "", MaxAge: -1, HttpOnly: false, Secure: secure, SameSite: http.SameSiteNoneMode, Path: "/"})
}

// RequireAuth validates the access_token cookie JWT.
// On failure, attempts a silent refresh via the refresh_token cookie.
func RequireAuth(authSvc services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try access token first
			cookie, err := c.Cookie("access_token")
			if err == nil && cookie.Value != "" {
				claims, err := authpkg.ValidateToken(cookie.Value)
				if err == nil {
					c.Set(AuthUserIDKey, claims.UserID)
					return next(c)
				}
			}

			// Attempt silent refresh
			refreshCookie, err := c.Cookie("refresh_token")
			if err != nil || refreshCookie.Value == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"message": "unauthorized",
				})
			}

			pair, err := authSvc.RefreshTokens(refreshCookie.Value)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"message": "unauthorized",
				})
			}

			claims, err := authpkg.ValidateToken(pair.AccessToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"message": "unauthorized",
				})
			}

			secure := CookieSecure()
			SetAccessCookie(c, pair.AccessToken, secure)
			SetRefreshCookie(c, pair.RefreshToken, secure)
			c.Set(AuthUserIDKey, claims.UserID)
			return next(c)
		}
	}
}
