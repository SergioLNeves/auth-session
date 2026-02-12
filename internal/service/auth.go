package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do"
	"go.uber.org/zap"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/pkg/logging"
)

type AuthServiceImpl struct {
	authRepository    domain.AuthRepository
	sessionRepository domain.SessionRepository
	tokenProvider     domain.TokenProvider
	passwordHasher    domain.PasswordHasher
}

func NewAuthService(i *do.Injector) (domain.AuthService, error) {
	authRepository := do.MustInvoke[domain.AuthRepository](i)
	sessionRepository := do.MustInvoke[domain.SessionRepository](i)
	tokenProvider := do.MustInvoke[domain.TokenProvider](i)
	passwordHasher := do.MustInvoke[domain.PasswordHasher](i)
	return &AuthServiceImpl{
		authRepository:    authRepository,
		sessionRepository: sessionRepository,
		tokenProvider:     tokenProvider,
		passwordHasher:    passwordHasher,
	}, nil
}

func (s *AuthServiceImpl) CreateAccount(ctx context.Context, req domain.CreateAccountRequest) (*domain.AuthResponse, error) {
	_, err := s.authRepository.FindUserByEmail(ctx, req.Email)
	if !errors.Is(err, domain.ErrUserNotFound) {
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		return nil, domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		ID:       uuid.New(),
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Avatar:   req.Avatar,
	}

	if err := s.authRepository.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(config.Env.Token.RefreshTokenExpiry) * time.Minute),
	}

	if err := s.sessionRepository.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	accessToken, err := s.tokenProvider.GenerateAccessToken(session.ID.String())
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

func (s *AuthServiceImpl) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.authRepository.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeactivated
	}

	if checkErr := s.passwordHasher.Check(req.Password, user.Password); checkErr != nil {
		return nil, domain.ErrInvalidCredentials
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(config.Env.Token.RefreshTokenExpiry) * time.Minute),
	}

	if createErr := s.sessionRepository.CreateSession(ctx, session); createErr != nil {
		return nil, fmt.Errorf("failed to create session: %w", createErr)
	}

	accessToken, err := s.tokenProvider.GenerateAccessToken(session.ID.String())
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

	session, err := s.sessionRepository.DeleteSession(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	logging.With(zap.String("service", "AuthService.Logout")).
		Info("session deleted",
			zap.String("session_id", session.ID.String()),
			zap.String("user_id", session.UserID.String()),
		)

	return nil
}

func (s *AuthServiceImpl) UpdatePassword(ctx context.Context, userID string, req domain.UpdatePasswordRequest) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.authRepository.FindUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if checkErr := s.passwordHasher.Check(req.CurrentPassword, user.Password); checkErr != nil {
		return domain.ErrInvalidCurrentPassword
	}

	hashedPassword, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = hashedPassword

	if err := s.authRepository.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *AuthServiceImpl) UpdateUser(ctx context.Context, userID string, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.authRepository.FindUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if req.Email != "" && req.Email != user.Email {
		_, findErr := s.authRepository.FindUserByEmail(ctx, req.Email)
		if !errors.Is(findErr, domain.ErrUserNotFound) {
			if findErr != nil {
				return nil, fmt.Errorf("failed to check email availability: %w", findErr)
			}
			return nil, domain.ErrEmailAlreadyExists
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := s.authRepository.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &domain.UserResponse{
		ID:     user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Avatar: user.Avatar,
	}, nil
}

func (s *AuthServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	if err := s.sessionRepository.DeleteSessionsByUserID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	if err := s.authRepository.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logging.With(zap.String("service", "AuthService.DeleteUser")).
		Info("user deactivated", zap.String("user_id", userID))

	return nil
}

func (s *AuthServiceImpl) ReactivateAccount(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.authRepository.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.DeletedAt == nil {
		return nil, domain.ErrUserNotDeactivated
	}

	if checkErr := s.passwordHasher.Check(req.Password, user.Password); checkErr != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user.DeletedAt = nil
	if updateErr := s.authRepository.UpdateUser(ctx, user); updateErr != nil {
		return nil, fmt.Errorf("failed to reactivate user: %w", updateErr)
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(config.Env.Token.RefreshTokenExpiry) * time.Minute),
	}

	if createErr := s.sessionRepository.CreateSession(ctx, session); createErr != nil {
		return nil, fmt.Errorf("failed to create session: %w", createErr)
	}

	accessToken, err := s.tokenProvider.GenerateAccessToken(session.ID.String())
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
