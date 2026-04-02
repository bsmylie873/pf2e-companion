package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CSRF returns Echo's CSRF middleware configured for cookie-based tokens.
// Auth endpoints (/auth/*) are skipped since clients won't have a CSRF token
// on their first visit (login/register).
func CSRF() echo.MiddlewareFunc {
	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/auth/")
		},
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteNoneMode,
		CookieSecure:   CookieSecure(),
		CookiePath:     "/",
	})
}
