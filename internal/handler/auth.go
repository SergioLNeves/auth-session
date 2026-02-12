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
	logger := logging.With(zap.String("handler", "AuthHandler.Login"))

	var request domain.LoginRequest
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

	response, err := e.AuthService.Login(c.Request().Context(), request)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			logger.Info("invalid credentials", zap.String("email", request.Email))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("auth", "invalid-credentials").
				WithTitle("Invalid Credentials").
				WithStatus(http.StatusUnauthorized).
				WithDetail("Invalid email or password").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusUnauthorized, problemDetails)
		}

		if errors.Is(err, domain.ErrUserDeactivated) {
			logger.Info("user deactivated", zap.String("email", request.Email))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("auth", "user-deactivated").
				WithTitle("Account Deactivated").
				WithStatus(http.StatusForbidden).
				WithDetail("Your account has been deactivated").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusForbidden, problemDetails)
		}

		logger.Error("failed to login", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("auth", "internal-error").
			WithTitle("Internal Server Error").
			WithStatus(http.StatusInternalServerError).
			WithDetail("An unexpected error occurred during login").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusInternalServerError, problemDetails)
	}

	setAuthCookies(c, response)

	return c.JSON(http.StatusOK, response)
}

func (e AuthHandlerImpl) Logout(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.Logout"))

	sessionID := c.Get("session_id").(string)
	if err := e.AuthService.Logout(c.Request().Context(), sessionID); err != nil {
		logger.Error("failed to deactivate session", zap.Error(err))
	}

	clearAuthCookies(c)
	return c.NoContent(http.StatusOK)
}

func (e AuthHandlerImpl) UpdatePassword(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.UpdatePassword"))

	var request domain.UpdatePasswordRequest
	if err := c.Bind(&request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "invalid-request").
			WithTitle("Invalid Request").
			WithStatus(http.StatusBadRequest).
			WithDetail("Failed to parse request body").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusBadRequest, problemDetails)
	}

	if err := validatorpkg.NewValidator().Validate(request); err != nil {
		logger.Error("validation failed", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "validation-error").
			WithTitle("Validation Failed").
			WithStatus(http.StatusBadRequest).
			WithDetail("One or more fields failed validation").
			WithInstance(c.Request().URL.Path).
			AddFieldErrors(
				errorpkg.NewProblemDetailsFromStructValidation(err.(validator.ValidationErrors)),
			)
		return c.JSON(http.StatusBadRequest, problemDetails)
	}

	userID := c.Get("user_id").(string)

	if err := e.AuthService.UpdatePassword(c.Request().Context(), userID, request); err != nil {
		if errors.Is(err, domain.ErrInvalidCurrentPassword) {
			logger.Info("invalid current password", zap.String("user_id", userID))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("user", "invalid-current-password").
				WithTitle("Invalid Current Password").
				WithStatus(http.StatusUnauthorized).
				WithDetail("The current password provided is incorrect").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusUnauthorized, problemDetails)
		}

		logger.Error("failed to update password", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "internal-error").
			WithTitle("Internal Server Error").
			WithStatus(http.StatusInternalServerError).
			WithDetail("An unexpected error occurred while updating the password").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusInternalServerError, problemDetails)
	}

	return c.NoContent(http.StatusNoContent)
}

func (e AuthHandlerImpl) UpdateUser(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.UpdateUser"))

	var request domain.UpdateUserRequest
	if err := c.Bind(&request); err != nil {
		logger.Error("failed to bind request", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "invalid-request").
			WithTitle("Invalid Request").
			WithStatus(http.StatusBadRequest).
			WithDetail("Failed to parse request body").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusBadRequest, problemDetails)
	}

	if err := validatorpkg.NewValidator().Validate(request); err != nil {
		logger.Error("validation failed", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "validation-error").
			WithTitle("Validation Failed").
			WithStatus(http.StatusBadRequest).
			WithDetail("One or more fields failed validation").
			WithInstance(c.Request().URL.Path).
			AddFieldErrors(
				errorpkg.NewProblemDetailsFromStructValidation(err.(validator.ValidationErrors)),
			)
		return c.JSON(http.StatusBadRequest, problemDetails)
	}

	userID := c.Get("user_id").(string)

	response, err := e.AuthService.UpdateUser(c.Request().Context(), userID, request)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			logger.Info("email already exists", zap.String("user_id", userID))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("user", "email-already-exists").
				WithTitle("Email Already Registered").
				WithStatus(http.StatusConflict).
				WithDetail("An account with this email already exists").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusConflict, problemDetails)
		}

		logger.Error("failed to update user", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "internal-error").
			WithTitle("Internal Server Error").
			WithStatus(http.StatusInternalServerError).
			WithDetail("An unexpected error occurred while updating the profile").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusInternalServerError, problemDetails)
	}

	return c.JSON(http.StatusOK, response)
}

func (e AuthHandlerImpl) Me(c echo.Context) error {
	return c.JSON(http.StatusOK, domain.UserResponse{
		ID:     c.Get("user_id").(string),
		Name:   c.Get("name").(string),
		Email:  c.Get("email").(string),
		Avatar: c.Get("avatar").(string),
	})
}

func (e AuthHandlerImpl) DeleteUser(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.DeleteUser"))

	userID := c.Get("user_id").(string)

	if err := e.AuthService.DeleteUser(c.Request().Context(), userID); err != nil {
		logger.Error("failed to delete user", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("user", "internal-error").
			WithTitle("Internal Server Error").
			WithStatus(http.StatusInternalServerError).
			WithDetail("An unexpected error occurred while deleting the account").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusInternalServerError, problemDetails)
	}

	clearAuthCookies(c)
	return c.NoContent(http.StatusOK)
}

func (e AuthHandlerImpl) ReactivateAccount(c echo.Context) error {
	logger := logging.With(zap.String("handler", "AuthHandler.ReactivateAccount"))

	var request domain.LoginRequest
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

	response, err := e.AuthService.ReactivateAccount(c.Request().Context(), request)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			logger.Info("invalid credentials", zap.String("email", request.Email))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("auth", "invalid-credentials").
				WithTitle("Invalid Credentials").
				WithStatus(http.StatusUnauthorized).
				WithDetail("Invalid email or password").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusUnauthorized, problemDetails)
		}

		if errors.Is(err, domain.ErrUserNotDeactivated) {
			logger.Info("user not deactivated", zap.String("email", request.Email))
			problemDetails := errorpkg.NewProblemDetails().
				WithType("auth", "user-not-deactivated").
				WithTitle("Account Not Deactivated").
				WithStatus(http.StatusBadRequest).
				WithDetail("This account is not deactivated").
				WithInstance(c.Request().URL.Path)
			return c.JSON(http.StatusBadRequest, problemDetails)
		}

		logger.Error("failed to reactivate account", zap.Error(err))
		problemDetails := errorpkg.NewProblemDetails().
			WithType("auth", "internal-error").
			WithTitle("Internal Server Error").
			WithStatus(http.StatusInternalServerError).
			WithDetail("An unexpected error occurred while reactivating the account").
			WithInstance(c.Request().URL.Path)
		return c.JSON(http.StatusInternalServerError, problemDetails)
	}

	setAuthCookies(c, response)

	return c.JSON(http.StatusOK, response)
}

func clearAuthCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
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
		HttpOnly: true,
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
