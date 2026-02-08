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
	tokenProvider  domain.TokenProvider
}

func NewAuthService(i *do.Injector) (domain.AuthService, error) {
	authRepository := do.MustInvoke[domain.AuthRepository](i)
	tokenProvider := do.MustInvoke[domain.TokenProvider](i)
	return &AuthServiceImpl{
		authRepository: authRepository,
		tokenProvider:  tokenProvider,
	}, nil
}

func (s *AuthServiceImpl) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.AuthResponse, error) {
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

	accessToken, err := s.tokenProvider.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenProvider.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
