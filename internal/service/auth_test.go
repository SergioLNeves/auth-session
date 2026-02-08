package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/pkg/logging"
	mockpkg "github.com/SergioLNeves/auth-session/mock"
)

func TestMain(m *testing.M) {
	logging.NewLogger(&domain.Config{Env: "development", LogLevel: "error"})
	os.Exit(m.Run())
}

func newAuthService(t *testing.T) (*AuthServiceImpl, *mockpkg.MockAuthRepository, *mockpkg.MockSessionRepository, *mockpkg.MockTokenProvider, *mockpkg.MockPasswordHasher) {
	t.Helper()
	authRepo := mockpkg.NewMockAuthRepository(t)
	sessionRepo := mockpkg.NewMockSessionRepository(t)
	tokenProvider := mockpkg.NewMockTokenProvider(t)
	passwordHasher := mockpkg.NewMockPasswordHasher(t)
	svc := &AuthServiceImpl{
		authRepository:    authRepo,
		sessionRepository: sessionRepo,
		tokenProvider:     tokenProvider,
		passwordHasher:    passwordHasher,
	}
	return svc, authRepo, sessionRepo, tokenProvider, passwordHasher
}

func TestCreateAccount(t *testing.T) {
	t.Run("should create account and return tokens", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, sessionRepo, tokenProvider, passwordHasher := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, nil)
		passwordHasher.On("Hash", "password123").Return("hashed-password", nil)
		authRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		sessionRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
		tokenProvider.On("GenerateAccessToken", mock.AnythingOfType("string"), "user@test.com", mock.AnythingOfType("string")).Return("access-token", nil)
		tokenProvider.On("GenerateRefreshToken", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("refresh-token", nil)

		result, err := svc.CreateAccount(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "access-token", result.AccessToken)
		assert.Equal(t, "refresh-token", result.RefreshToken)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, _ := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}
		existingUser := &domain.User{ID: uuid.New(), Email: "user@test.com"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(existingUser, nil)

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})

	t.Run("should return error when FindUserByEmail fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, _ := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, errors.New("db error"))

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check existing email")
	})

	t.Run("should return error when Hash fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, passwordHasher := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, nil)
		passwordHasher.On("Hash", "password123").Return("", errors.New("hash error"))

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to hash password")
	})

	t.Run("should return error when CreateUser fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, passwordHasher := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, nil)
		passwordHasher.On("Hash", "password123").Return("hashed-password", nil)
		authRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db error"))

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
	})

	t.Run("should return error when CreateSession fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, sessionRepo, _, passwordHasher := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, nil)
		passwordHasher.On("Hash", "password123").Return("hashed-password", nil)
		authRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		sessionRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(errors.New("db error"))

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create session")
	})

	t.Run("should return error when GenerateAccessToken fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, sessionRepo, tokenProvider, passwordHasher := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, nil)
		passwordHasher.On("Hash", "password123").Return("hashed-password", nil)
		authRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		sessionRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
		tokenProvider.On("GenerateAccessToken", mock.AnythingOfType("string"), "user@test.com", mock.AnythingOfType("string")).Return("", errors.New("token error"))

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate access token")
	})

	t.Run("should return error when GenerateRefreshToken fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, sessionRepo, tokenProvider, passwordHasher := newAuthService(t)
		ctx := context.Background()
		req := domain.CreateAccountRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, nil)
		passwordHasher.On("Hash", "password123").Return("hashed-password", nil)
		authRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).Return(nil)
		sessionRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
		tokenProvider.On("GenerateAccessToken", mock.AnythingOfType("string"), "user@test.com", mock.AnythingOfType("string")).Return("access-token", nil)
		tokenProvider.On("GenerateRefreshToken", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", errors.New("token error"))

		result, err := svc.CreateAccount(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate refresh token")
	})
}

func TestLogin(t *testing.T) {
	t.Run("should login and return tokens", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, sessionRepo, tokenProvider, passwordHasher := newAuthService(t)
		ctx := context.Background()
		user := &domain.User{ID: uuid.New(), Email: "user@test.com", Password: "hashed-password"}
		req := domain.LoginRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(user, nil)
		passwordHasher.On("Check", "password123", "hashed-password").Return(nil)
		sessionRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
		tokenProvider.On("GenerateAccessToken", user.ID.String(), "user@test.com", mock.AnythingOfType("string")).Return("access-token", nil)
		tokenProvider.On("GenerateRefreshToken", user.ID.String(), mock.AnythingOfType("string")).Return("refresh-token", nil)

		result, err := svc.Login(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "access-token", result.AccessToken)
		assert.Equal(t, "refresh-token", result.RefreshToken)
	})

	t.Run("should return ErrInvalidCredentials when user not found", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, _ := newAuthService(t)
		ctx := context.Background()
		req := domain.LoginRequest{Email: "nobody@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "nobody@test.com").Return(nil, nil)

		result, err := svc.Login(ctx, req)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("should return ErrInvalidCredentials when password is wrong", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, passwordHasher := newAuthService(t)
		ctx := context.Background()
		user := &domain.User{ID: uuid.New(), Email: "user@test.com", Password: "hashed-password"}
		req := domain.LoginRequest{Email: "user@test.com", Password: "wrongpassword"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(user, nil)
		passwordHasher.On("Check", "wrongpassword", "hashed-password").Return(errors.New("mismatch"))

		result, err := svc.Login(ctx, req)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("should return error when FindUserByEmail fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, _, _, _ := newAuthService(t)
		ctx := context.Background()
		req := domain.LoginRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(nil, errors.New("db error"))

		result, err := svc.Login(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find user")
	})

	t.Run("should return error when CreateSession fails", func(t *testing.T) {
		t.Parallel()

		svc, authRepo, sessionRepo, _, passwordHasher := newAuthService(t)
		ctx := context.Background()
		user := &domain.User{ID: uuid.New(), Email: "user@test.com", Password: "hashed-password"}
		req := domain.LoginRequest{Email: "user@test.com", Password: "password123"}

		authRepo.On("FindUserByEmail", ctx, "user@test.com").Return(user, nil)
		passwordHasher.On("Check", "password123", "hashed-password").Return(nil)
		sessionRepo.On("CreateSession", ctx, mock.AnythingOfType("*domain.Session")).Return(errors.New("db error"))

		result, err := svc.Login(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create session")
	})
}

func TestLogout(t *testing.T) {
	t.Run("should delete session successfully", func(t *testing.T) {
		t.Parallel()

		svc, _, sessionRepo, _, _ := newAuthService(t)
		ctx := context.Background()
		sessionID := uuid.New()
		deletedSession := &domain.Session{ID: sessionID, UserID: uuid.New()}

		sessionRepo.On("DeleteSession", ctx, sessionID).Return(deletedSession, nil)

		err := svc.Logout(ctx, sessionID.String())

		assert.NoError(t, err)
	})

	t.Run("should return error on invalid session ID format", func(t *testing.T) {
		t.Parallel()

		svc, _, _, _, _ := newAuthService(t)
		ctx := context.Background()

		err := svc.Logout(ctx, "not-a-uuid")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid session ID")
	})

	t.Run("should return error when DeleteSession fails", func(t *testing.T) {
		t.Parallel()

		svc, _, sessionRepo, _, _ := newAuthService(t)
		ctx := context.Background()
		sessionID := uuid.New()

		sessionRepo.On("DeleteSession", ctx, sessionID).Return(nil, errors.New("db error"))

		err := svc.Logout(ctx, sessionID.String())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete session")
	})
}
