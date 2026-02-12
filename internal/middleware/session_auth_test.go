package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

func newMiddlewareContext(accessToken, refreshToken string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	if accessToken != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	}
	if refreshToken != "" {
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func dummyNext(_ echo.Context) error {
	return nil
}

func TestSessionAuth(t *testing.T) {
	t.Run("should return 401 when access_token cookie is missing", func(t *testing.T) {
		t.Parallel()

		tokenProvider := mockpkg.NewMockTokenProvider(t)
		sessionRepo := mockpkg.NewMockSessionRepository(t)
		authRepo := mockpkg.NewMockAuthRepository(t)

		c, rec := newMiddlewareContext("", "")
		handler := SessionAuth(tokenProvider, sessionRepo, authRepo)(dummyNext)

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 401 when access token is invalid", func(t *testing.T) {
		t.Parallel()

		tokenProvider := mockpkg.NewMockTokenProvider(t)
		sessionRepo := mockpkg.NewMockSessionRepository(t)
		authRepo := mockpkg.NewMockAuthRepository(t)

		tokenProvider.On("ParseAccessToken", "bad-token").Return(nil, errors.New("invalid"))

		c, rec := newMiddlewareContext("bad-token", "")
		handler := SessionAuth(tokenProvider, sessionRepo, authRepo)(dummyNext)

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 401 when session not found", func(t *testing.T) {
		t.Parallel()

		tokenProvider := mockpkg.NewMockTokenProvider(t)
		sessionRepo := mockpkg.NewMockSessionRepository(t)
		authRepo := mockpkg.NewMockAuthRepository(t)

		sessionID := uuid.New()
		claims := &domain.AccessTokenClaims{SessionID: sessionID.String()}
		tokenProvider.On("ParseAccessToken", "valid-token").Return(claims, nil)
		sessionRepo.On("FindSessionByID", mock.Anything, sessionID).Return(nil, domain.ErrSessionNotFound)

		c, rec := newMiddlewareContext("valid-token", "")
		handler := SessionAuth(tokenProvider, sessionRepo, authRepo)(dummyNext)

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 401 when refresh token is missing", func(t *testing.T) {
		t.Parallel()

		tokenProvider := mockpkg.NewMockTokenProvider(t)
		sessionRepo := mockpkg.NewMockSessionRepository(t)
		authRepo := mockpkg.NewMockAuthRepository(t)

		sessionID := uuid.New()
		session := &domain.Session{ID: sessionID, UserID: uuid.New()}
		claims := &domain.AccessTokenClaims{SessionID: sessionID.String()}
		tokenProvider.On("ParseAccessToken", "valid-token").Return(claims, nil)
		sessionRepo.On("FindSessionByID", mock.Anything, sessionID).Return(session, nil)

		c, rec := newMiddlewareContext("valid-token", "")
		handler := SessionAuth(tokenProvider, sessionRepo, authRepo)(dummyNext)

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should delete session and return 401 when refresh token is expired", func(t *testing.T) {
		t.Parallel()

		tokenProvider := mockpkg.NewMockTokenProvider(t)
		sessionRepo := mockpkg.NewMockSessionRepository(t)
		authRepo := mockpkg.NewMockAuthRepository(t)

		sessionID := uuid.New()
		session := &domain.Session{ID: sessionID, UserID: uuid.New()}
		claims := &domain.AccessTokenClaims{SessionID: sessionID.String()}
		tokenProvider.On("ParseAccessToken", "valid-token").Return(claims, nil)
		sessionRepo.On("FindSessionByID", mock.Anything, sessionID).Return(session, nil)
		tokenProvider.On("ParseRefreshToken", "expired-refresh").Return(nil, errors.New("expired"))
		sessionRepo.On("DeleteSession", mock.Anything, sessionID).Return(session, nil)

		c, rec := newMiddlewareContext("valid-token", "expired-refresh")
		handler := SessionAuth(tokenProvider, sessionRepo, authRepo)(dummyNext)

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should pass through and set context on valid tokens", func(t *testing.T) {
		t.Parallel()

		tokenProvider := mockpkg.NewMockTokenProvider(t)
		sessionRepo := mockpkg.NewMockSessionRepository(t)
		authRepo := mockpkg.NewMockAuthRepository(t)

		userID := uuid.New()
		sessionID := uuid.New()
		session := &domain.Session{ID: sessionID, UserID: userID}
		user := &domain.User{ID: userID, Email: "user@test.com"}
		accessClaims := &domain.AccessTokenClaims{SessionID: sessionID.String()}
		refreshClaims := &domain.RefreshTokenClaims{UserID: userID.String(), SessionID: sessionID.String()}

		tokenProvider.On("ParseAccessToken", "valid-token").Return(accessClaims, nil)
		sessionRepo.On("FindSessionByID", mock.Anything, sessionID).Return(session, nil)
		tokenProvider.On("ParseRefreshToken", "valid-refresh").Return(refreshClaims, nil)
		authRepo.On("FindUserByID", mock.Anything, userID).Return(user, nil)
		tokenProvider.On("GenerateAccessToken", sessionID.String()).Return("new-access", nil)
		tokenProvider.On("GenerateRefreshToken", userID.String(), sessionID.String()).Return("new-refresh", nil)
		sessionRepo.On("UpdateSessionExpiry", mock.Anything, sessionID, mock.AnythingOfType("time.Time")).Return(nil)

		var ctxUserID, ctxEmail, ctxSessionID string
		next := func(c echo.Context) error {
			ctxUserID = c.Get("user_id").(string)
			ctxEmail = c.Get("email").(string)
			ctxSessionID = c.Get("session_id").(string)
			return nil
		}

		c, rec := newMiddlewareContext("valid-token", "valid-refresh")
		handler := SessionAuth(tokenProvider, sessionRepo, authRepo)(next)

		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, userID.String(), ctxUserID)
		assert.Equal(t, "user@test.com", ctxEmail)
		assert.Equal(t, sessionID.String(), ctxSessionID)
	})
}
