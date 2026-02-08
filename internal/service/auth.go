package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/security"
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

func (s *AuthServiceImpl) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.User, error) {
	existingUser, err := s.authRepository.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: *hashedPassword,
		Active:   true,
	}

	if err := s.authRepository.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
