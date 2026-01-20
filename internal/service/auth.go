package service

import (
	"context"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/samber/do"
)

type AuthServiceImpl struct {
	authRepository domain.AuthRepository
}

func NewAuthService(i *do.Injector) (domain.AuthService, error) {
	authRepository := do.MustInvoke[domain.AuthRepository](i)
	return &AuthServiceImpl{
		authRepository: authRepository,
	}, nil
}

func (s AuthServiceImpl) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.User, error) {
	return nil, nil
}
