package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	errorpkg "github.com/SergioLNeves/auth-session/internal/pkg/error"
	"github.com/SergioLNeves/auth-session/internal/pkg/logging"
)

func SessionAuth(
	tokenProvider domain.TokenProvider,
	sessionRepo domain.SessionRepository,
	authRepo domain.AuthRepository,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := logging.With(zap.String("middleware", "SessionAuth"))

			cookie, err := c.Cookie("access_token")
			if err != nil || cookie.Value == "" {
				logger.Warn("missing access token cookie")
				return unauthorizedResponse(c)
			}

			claims, err := tokenProvider.ParseAccessToken(cookie.Value)
			if err != nil {
				logger.Warn("invalid access token", zap.Error(err))
				return unauthorizedResponse(c)
			}

			sessionID, err := uuid.Parse(claims.SessionID)
			if err != nil {
				logger.Warn("invalid session ID in token", zap.Error(err))
				return unauthorizedResponse(c)
			}

			session, err := sessionRepo.FindSessionByID(c.Request().Context(), sessionID)
			if err != nil {
				logger.Error("failed to find session", zap.Error(err))
				return internalErrorResponse(c)
			}

			if session == nil || !session.Active {
				logger.Info("session not active", zap.String("session_id", sessionID.String()))
				clearAuthCookies(c)
				return unauthorizedResponse(c)
			}

			// Try refresh flow: parse refresh token to check if it's still valid
			refreshCookie, err := c.Cookie("refresh_token")
			if err != nil || refreshCookie.Value == "" {
				logger.Warn("missing refresh token cookie")
				clearAuthCookies(c)
				return unauthorizedResponse(c)
			}

			_, refreshErr := tokenProvider.ParseRefreshToken(refreshCookie.Value)
			if refreshErr != nil {
				logger.Info("refresh token expired, clearing session", zap.Error(refreshErr))
				if deactivateErr := sessionRepo.DeactivateSession(c.Request().Context(), sessionID); deactivateErr != nil {
					logger.Error("failed to deactivate expired session", zap.Error(deactivateErr))
				}
				clearAuthCookies(c)
				return unauthorizedResponse(c)
			}

			// Refresh token is valid — regenerate tokens
			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				logger.Warn("invalid user ID in token", zap.Error(err))
				return unauthorizedResponse(c)
			}

			user, err := authRepo.FindUserByID(c.Request().Context(), userID)
			if err != nil || user == nil {
				logger.Error("failed to find user for token refresh", zap.Error(err))
				return internalErrorResponse(c)
			}

			newAccessToken, err := tokenProvider.GenerateAccessToken(user.ID.String(), user.Email, session.ID.String())
			if err != nil {
				logger.Error("failed to generate new access token", zap.Error(err))
				return internalErrorResponse(c)
			}

			newRefreshToken, err := tokenProvider.GenerateRefreshToken(user.ID.String(), session.ID.String())
			if err != nil {
				logger.Error("failed to generate new refresh token", zap.Error(err))
				return internalErrorResponse(c)
			}

			setAuthCookies(c, &domain.AuthResponse{
				AccessToken:  newAccessToken,
				RefreshToken: newRefreshToken,
			})

			c.Set("user_id", user.ID.String())
			c.Set("email", user.Email)
			c.Set("session_id", session.ID.String())

			return next(c)
		}
	}
}

func clearAuthCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
	})
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func setAuthCookies(c echo.Context, response *domain.AuthResponse) {
	isProduction := config.Env.Env == "production"

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    response.AccessToken,
		Path:     "/",
		MaxAge:   config.Env.Token.AccessTokenExpiry * 60,
		HttpOnly: false,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
	})

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    response.RefreshToken,
		Path:     "/",
		MaxAge:   config.Env.Token.RefreshTokenExpiry * 60,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
	})
}

func unauthorizedResponse(c echo.Context) error {
	problemDetails := errorpkg.NewProblemDetails().
		WithType("auth", "unauthorized").
		WithTitle("Unauthorized").
		WithStatus(http.StatusUnauthorized).
		WithDetail("Authentication required").
		WithInstance(c.Request().URL.Path)
	return c.JSON(http.StatusUnauthorized, problemDetails)
}

func internalErrorResponse(c echo.Context) error {
	problemDetails := errorpkg.NewProblemDetails().
		WithType("auth", "internal-error").
		WithTitle("Internal Server Error").
		WithStatus(http.StatusInternalServerError).
		WithDetail("An unexpected error occurred").
		WithInstance(c.Request().URL.Path)
	return c.JSON(http.StatusInternalServerError, problemDetails)
}
