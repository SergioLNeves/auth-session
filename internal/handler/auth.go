package handler

import (
	"net/http"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
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
	return c.NoContent(http.StatusOK)
}

func (e AuthHandlerImpl) Login(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
