package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// BackupRateLimiter returns a middleware that limits backup operations to 10 per user per hour.
func BackupRateLimiter() echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Skipper: echomw.DefaultSkipper,
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(10.0 / 3600.0),
				Burst:     10,
				ExpiresIn: 1 * time.Hour,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			val := ctx.Get(AuthUserIDKey)
			if val == nil {
				return ctx.RealIP(), nil
			}
			return val.(uuid.UUID).String(), nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    http.StatusForbidden,
				"message": "forbidden",
			})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    http.StatusTooManyRequests,
				"message": "backup rate limit exceeded (10 per hour)",
			})
		},
	})
}

// RateLimiter returns a rate limiting middleware limiting each IP to 60 req/min.
func RateLimiter() echo.MiddlewareFunc {
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Skipper: echomw.DefaultSkipper,
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(
			echomw.RateLimiterMemoryStoreConfig{
				Rate:      60,
				Burst:     10,
				ExpiresIn: 3 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, map[string]interface{}{
				"code":    http.StatusForbidden,
				"message": "forbidden",
			})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"code":    http.StatusTooManyRequests,
				"message": "too many requests",
			})
		},
	})
}
