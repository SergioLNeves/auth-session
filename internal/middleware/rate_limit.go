package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	errorpkg "github.com/SergioLNeves/auth-session/internal/pkg/error"
)

// NewRateLimiter cria um middleware de rate limit por IP.
// rpm = requisições por minuto permitidas; burst = pico máximo.
func NewRateLimiter(rpm, burst int) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:  rate.Limit(float64(rpm) / 60.0),
				Burst: burst,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			problemDetails := errorpkg.NewProblemDetails().
				WithType("rate-limit", "too-many-requests").
				WithTitle("Too Many Requests").
				WithStatus(http.StatusTooManyRequests).
				WithDetail("Rate limit exceeded. Please try again later.").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusTooManyRequests, problemDetails)
		},
	})
}
