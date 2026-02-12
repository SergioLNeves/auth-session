package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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

func newHandler(t *testing.T) (*AuthHandlerImpl, *mockpkg.MockAuthService) {
	t.Helper()
	authService := mockpkg.NewMockAuthService(t)
	h := &AuthHandlerImpl{AuthService: authService}
	return h, authService
}

func newFormContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestCreateAccount(t *testing.T) {
	t.Run("should return 201 and tokens on success", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/user/create-account", "name=Test+User&email=user@test.com&password=password123")

		authService.On("CreateAccount", mock.Anything, domain.CreateAccountRequest{
			Name: "Test User", Email: "user@test.com", Password: "password123",
		}).Return(&domain.AuthResponse{AccessToken: "at", RefreshToken: "rt"}, nil)

		err := h.CreateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp domain.AuthResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "at", resp.AccessToken)
		assert.Equal(t, "rt", resp.RefreshToken)
	})

	t.Run("should return 409 when email already exists", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/user/create-account", "name=Test+User&email=user@test.com&password=password123")

		authService.On("CreateAccount", mock.Anything, domain.CreateAccountRequest{
			Name: "Test User", Email: "user@test.com", Password: "password123",
		}).Return(nil, domain.ErrEmailAlreadyExists)

		err := h.CreateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("should return 500 on service error", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/user/create-account", "name=Test+User&email=user@test.com&password=password123")

		authService.On("CreateAccount", mock.Anything, domain.CreateAccountRequest{
			Name: "Test User", Email: "user@test.com", Password: "password123",
		}).Return(nil, errors.New("unexpected"))

		err := h.CreateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		t.Parallel()

		h, _ := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/user/create-account", "email=invalid&password=short")

		err := h.CreateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestLogin(t *testing.T) {
	t.Run("should return 200 and tokens on success", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/auth/login", "email=user@test.com&password=password123")

		authService.On("Login", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "password123",
		}).Return(&domain.AuthResponse{AccessToken: "at", RefreshToken: "rt"}, nil)

		err := h.Login(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp domain.AuthResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "at", resp.AccessToken)
		assert.Equal(t, "rt", resp.RefreshToken)
	})

	t.Run("should return 401 on invalid credentials", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/auth/login", "email=user@test.com&password=wrong")

		authService.On("Login", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "wrong",
		}).Return(nil, domain.ErrInvalidCredentials)

		err := h.Login(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 500 on service error", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/auth/login", "email=user@test.com&password=password123")

		authService.On("Login", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "password123",
		}).Return(nil, errors.New("unexpected"))

		err := h.Login(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("should return 403 when user is deactivated", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/auth/login", "email=user@test.com&password=password123")

		authService.On("Login", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "password123",
		}).Return(nil, domain.ErrUserDeactivated)

		err := h.Login(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		t.Parallel()

		h, _ := newHandler(t)
		c, rec := newFormContext(http.MethodPost, "/v1/auth/login", "email=invalid&password=")

		err := h.Login(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestLogout(t *testing.T) {
	t.Run("should return 200 and clear cookies", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("session_id", "some-session-id")

		authService.On("Logout", mock.Anything, "some-session-id").Return(nil)

		err := h.Logout(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("should still return 200 when service logout fails", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("session_id", "some-session-id")

		authService.On("Logout", mock.Anything, "some-session-id").Return(errors.New("db error"))

		err := h.Logout(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestUpdatePassword(t *testing.T) {
	t.Run("should return 204 on success", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/v1/user/password", strings.NewReader("current_password=oldpass123&new_password=newpass123"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("UpdatePassword", mock.Anything, "some-user-id", domain.UpdatePasswordRequest{
			CurrentPassword: "oldpass123", NewPassword: "newpass123",
		}).Return(nil)

		err := h.UpdatePassword(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		t.Parallel()

		h, _ := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/password", "")

		err := h.UpdatePassword(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("should return 401 when current password is wrong", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/v1/user/password", strings.NewReader("current_password=wrongpass&new_password=newpass123"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("UpdatePassword", mock.Anything, "some-user-id", domain.UpdatePasswordRequest{
			CurrentPassword: "wrongpass", NewPassword: "newpass123",
		}).Return(domain.ErrInvalidCurrentPassword)

		err := h.UpdatePassword(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 500 on service error", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/v1/user/password", strings.NewReader("current_password=oldpass123&new_password=newpass123"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("UpdatePassword", mock.Anything, "some-user-id", domain.UpdatePasswordRequest{
			CurrentPassword: "oldpass123", NewPassword: "newpass123",
		}).Return(errors.New("unexpected"))

		err := h.UpdatePassword(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestMe(t *testing.T) {
	t.Run("should return 200 with user data from context", func(t *testing.T) {
		t.Parallel()

		h, _ := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")
		c.Set("name", "Test User")
		c.Set("email", "user@test.com")
		c.Set("avatar", "https://example.com/avatar.png")

		err := h.Me(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp domain.UserResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "some-user-id", resp.ID)
		assert.Equal(t, "Test User", resp.Name)
		assert.Equal(t, "user@test.com", resp.Email)
		assert.Equal(t, "https://example.com/avatar.png", resp.Avatar)
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("should return 200 on success", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/v1/user", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("DeleteUser", mock.Anything, "some-user-id").Return(nil)

		err := h.DeleteUser(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("should return 500 on service error", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/v1/user", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("DeleteUser", mock.Anything, "some-user-id").Return(errors.New("unexpected"))

		err := h.DeleteUser(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("should return 200 with updated user on success", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/v1/user/profile", strings.NewReader("name=New+Name&email=new@test.com&avatar=http://avatar.com/pic.png"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("UpdateUser", mock.Anything, "some-user-id", domain.UpdateUserRequest{
			Name: "New Name", Email: "new@test.com", Avatar: "http://avatar.com/pic.png",
		}).Return(&domain.UserResponse{
			ID: "some-user-id", Name: "New Name", Email: "new@test.com", Avatar: "http://avatar.com/pic.png",
		}, nil)

		err := h.UpdateUser(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp domain.UserResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "some-user-id", resp.ID)
		assert.Equal(t, "New Name", resp.Name)
		assert.Equal(t, "new@test.com", resp.Email)
		assert.Equal(t, "http://avatar.com/pic.png", resp.Avatar)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		t.Parallel()

		h, _ := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/profile", "email=not-an-email")

		err := h.UpdateUser(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("should return 409 when email already exists", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/v1/user/profile", strings.NewReader("email=taken@test.com"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("UpdateUser", mock.Anything, "some-user-id", domain.UpdateUserRequest{
			Email: "taken@test.com",
		}).Return(nil, domain.ErrEmailAlreadyExists)

		err := h.UpdateUser(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("should return 500 on service error", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodPatch, "/v1/user/profile", strings.NewReader("name=New+Name"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_id", "some-user-id")

		authService.On("UpdateUser", mock.Anything, "some-user-id", domain.UpdateUserRequest{
			Name: "New Name",
		}).Return(nil, errors.New("unexpected"))

		err := h.UpdateUser(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestReactivateAccount(t *testing.T) {
	t.Run("should return 200 and tokens on success", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/reactivate", "email=user@test.com&password=password123")

		authService.On("ReactivateAccount", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "password123",
		}).Return(&domain.AuthResponse{AccessToken: "at", RefreshToken: "rt"}, nil)

		err := h.ReactivateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp domain.AuthResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "at", resp.AccessToken)
		assert.Equal(t, "rt", resp.RefreshToken)
	})

	t.Run("should return 400 on validation error", func(t *testing.T) {
		t.Parallel()

		h, _ := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/reactivate", "email=invalid&password=")

		err := h.ReactivateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("should return 401 on invalid credentials", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/reactivate", "email=user@test.com&password=wrong")

		authService.On("ReactivateAccount", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "wrong",
		}).Return(nil, domain.ErrInvalidCredentials)

		err := h.ReactivateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("should return 400 when user is not deactivated", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/reactivate", "email=user@test.com&password=password123")

		authService.On("ReactivateAccount", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "password123",
		}).Return(nil, domain.ErrUserNotDeactivated)

		err := h.ReactivateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("should return 500 on service error", func(t *testing.T) {
		t.Parallel()

		h, authService := newHandler(t)
		c, rec := newFormContext(http.MethodPatch, "/v1/user/reactivate", "email=user@test.com&password=password123")

		authService.On("ReactivateAccount", mock.Anything, domain.LoginRequest{
			Email: "user@test.com", Password: "password123",
		}).Return(nil, errors.New("unexpected"))

		err := h.ReactivateAccount(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
