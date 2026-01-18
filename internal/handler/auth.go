package handler

import (
	"net/http"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/labstack/echo/v4"
)

type AuthHandlerImpl struct {
	AuthService domain.AuthService
}

func NewAuthHandler(AuthService domain.AuthService) (domain.AuthHandler, error) {
	return &AuthHandlerImpl{
		AuthService: AuthService,
	}, nil
}

func (e AuthHandlerImpl) CreateAccount(ectx echo.Context) error {

	return ectx.NoContent(http.StatusOK)
}

func (e AuthHandlerImpl) Login(ectx echo.Context) error {

	return ectx.NoContent(http.StatusOK)
}
