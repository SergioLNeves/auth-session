package handler

import (
	"errors"
	"net/http"

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

	return c.JSON(http.StatusCreated, response)
}

func (e AuthHandlerImpl) Login(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
