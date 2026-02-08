package handler

import (
	"errors"
	"net/http"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	errorpkg "github.com/SergioLNeves/auth-session/internal/pkg/error"
	"github.com/SergioLNeves/auth-session/internal/pkg/logging"
	validatorpkg "github.com/SergioLNeves/auth-session/internal/pkg/validator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
	"go.uber.org/zap"
)

type AuthHandlerImpl struct {
	AuthService domain.AuthService
}

func NewAuthHandler(i *do.Injector) (domain.AuthHandler, error) {
	authService := do.MustInvoke[domain.AuthService](i)

	return &AuthHandlerImpl{
		AuthService: authService,
	}, nil
}

func (e AuthHandlerImpl) CreateAccount(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.CreateAccount"))

	var request domain.CreateAccountRequest
	if err := c.Bind(&request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("auth", "invalid-request").
			WithTitle("Invalid Request").
			WithStatus(http.StatusBadRequest).
			WithDetail("Failed to parse request body").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusBadRequest, problemDetails)
	}

	if err := validatorpkg.NewValidator().Validate(request); err != nil {
		logger.Error("validation failed", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("auth", "validation-error").
			WithTitle("Validation Failed").
			WithStatus(http.StatusBadRequest).
			WithDetail("One or more fields failed validation").
			WithInstance(c.Request().URL.Path).
			AddFieldErrors(
				errorpkg.NewProblemDetailsFromStructValidation(err.(validator.ValidationErrors)),
			)
		return c.JSON(http.StatusBadRequest, problemDetails)
	}

	response, err := e.AuthService.CreateAccount(c.Request().Context(), request)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			logger.Info("email already exists", zap.String("email", request.Email))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("auth", "email-already-exists").
				WithTitle("Email Already Registered").
				WithStatus(http.StatusConflict).
				WithDetail("An account with this email already exists").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusConflict, problemDetails)
		}

		logger.Error("failed to create account", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("auth", "internal-error").
			WithTitle("Internal Server Error").
			WithStatus(http.StatusInternalServerError).
			WithDetail("An unexpected error occurred while creating the account").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusInternalServerError, problemDetails)
	}

	setAuthCookies(c, response)

	return c.JSON(http.StatusCreated, response)
}

func (e AuthHandlerImpl) Login(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func (e AuthHandlerImpl) Logout(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.Logout"))

	cookie, err := c.Cookie("access_token")
	if err != nil || cookie.Value == "" {
		logger.Warn("logout attempt without access token cookie")
		clearAuthCookies(c)
		return c.NoContent(http.StatusOK)
	}

	if err := e.AuthService.Logout(c.Request().Context(), cookie.Value); err != nil {
		logger.Error("failed to deactivate session", zap.Error(err))
	}

	clearAuthCookies(c)
	return c.NoContent(http.StatusOK)
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

	// MaxAge expects seconds; env values are in minutes, so multiply by 60
	// access_token is readable by JS to extract user claims (email)
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
