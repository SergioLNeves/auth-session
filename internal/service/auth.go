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
	authRepository    domain.AuthRepository
	sessionRepository domain.SessionRepository
	tokenProvider     domain.TokenProvider
}

func NewAuthService(i *do.Injector) (domain.AuthService, error) {
	authRepository := do.MustInvoke[domain.AuthRepository](i)
	sessionRepository := do.MustInvoke[domain.SessionRepository](i)
	tokenProvider := do.MustInvoke[domain.TokenProvider](i)
	return &AuthServiceImpl{
		authRepository:    authRepository,
		sessionRepository: sessionRepository,
		tokenProvider:     tokenProvider,
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

	session := &domain.Session{
		ID:     uuid.New(),
		UserID: user.ID,
		Active: true,
	}

	if err := s.sessionRepository.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	accessToken, err := s.tokenProvider.GenerateAccessToken(user.ID.String(), user.Email, session.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenProvider.GenerateRefreshToken(user.ID.String(), session.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	if err := s.sessionRepository.DeactivateSession(ctx, id); err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	return nil
}
