package main

import (
	"context"
	"time"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/handler"
	authmiddleware "github.com/SergioLNeves/auth-session/internal/middleware"
	"github.com/SergioLNeves/auth-session/internal/pkg/logging"
	validator "github.com/SergioLNeves/auth-session/internal/pkg/validator"
	"github.com/SergioLNeves/auth-session/internal/repository"
	"github.com/SergioLNeves/auth-session/internal/security"
	"github.com/SergioLNeves/auth-session/internal/service"
	"github.com/SergioLNeves/auth-session/internal/storage/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
	"go.uber.org/zap"
)

var (
	injector *do.Injector
	logger   *zap.Logger
)

func main() {
	if err := config.LoadEnv(); err != nil {
		panic("failed to load environment: " + err.Error())
	}

	logger = logging.NewLogger(&config.Env)
	defer logger.Sync()

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Validator = validator.NewValidator()

	initDependencies(logger)
	defer func() {
		if err := injector.Shutdown(); err != nil {
			e.Logger.Errorf("shutdown injector: %w", err)
		}
	}()

	configureHealthcheckRoute(e)
	configureAuthRoute(e)

	sessionRepo := do.MustInvoke[domain.SessionRepository](injector)
	startSessionCleanup(sessionRepo)

	authRepo := do.MustInvoke[domain.AuthRepository](injector)
	startUserCleanup(authRepo)

	api := config.NewAPI(e, config.Env.Port, 10*time.Second)
	api.Start()
}

func configureHealthcheckRoute(e *echo.Echo) {
	healthCheckHandler, err := do.Invoke[domain.HealthCheckHandler](injector)
	if err != nil {
		logger.Fatal("invoke healthcheck handler", zap.Error(err))
	}

	e.GET("/health", healthCheckHandler.Check)
}

func configureAuthRoute(e *echo.Echo) {
	tokenProvider := do.MustInvoke[domain.TokenProvider](injector)
	sessionRepo := do.MustInvoke[domain.SessionRepository](injector)
	authRepo := do.MustInvoke[domain.AuthRepository](injector)
	authHandler, err := do.Invoke[domain.AuthHandler](injector)
	if err != nil {
		logger.Fatal("invoke auth handler", zap.Error(err))
	}
	sessionAuth := authmiddleware.SessionAuth(tokenProvider, sessionRepo, authRepo)

	v1 := e.Group("/v1")
	userGroup := v1.Group("/user")
	userGroup.POST("/create-account", authHandler.CreateAccount)
	userGroup.PATCH("/password", authHandler.UpdatePassword, sessionAuth)
	userGroup.PATCH("/profile", authHandler.UpdateUser, sessionAuth)
	userGroup.DELETE("", authHandler.DeleteUser, sessionAuth)
	userGroup.PATCH("/reactivate", authHandler.ReactivateAccount)

	authGroup := v1.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/logout", authHandler.Logout, sessionAuth)
	authGroup.GET("/me", authHandler.Me, sessionAuth)
}

func startSessionCleanup(sessionRepo domain.SessionRepository) {
	ticker := time.NewTicker(12 * time.Hour)
	go func() {
		for range ticker.C {
			deleted, err := sessionRepo.DeleteExpiredSessions(context.Background())
			if err != nil {
				logger.Error("session cleanup failed", zap.Error(err))
				continue
			}
			if deleted > 0 {
				logger.Info("expired sessions cleaned up", zap.Int64("deleted", deleted))
			}
		}
	}()
}

func startUserCleanup(authRepo domain.AuthRepository) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			deleted, err := authRepo.DeleteDeactivatedUsers(context.Background())
			if err != nil {
				logger.Error("user cleanup failed", zap.Error(err))
				continue
			}
			if deleted > 0 {
				logger.Info("deactivated users cleaned up", zap.Int64("deleted", deleted))
			}
		}
	}()
}

func initDependencies(logger *zap.Logger) {
	injector = do.New()

	do.ProvideValue(injector, logger)

	do.Provide(injector, sqlite.NewSQLite)

	do.Provide(injector, repository.NewAuthRepository)
	do.Provide(injector, repository.NewSessionRepository)

	do.Provide(injector, security.NewJWTProvider)
	do.Provide(injector, security.NewBcryptHasher)

	do.Provide(injector, service.NewHealthCheckService)
	do.Provide(injector, service.NewAuthService)

	do.Provide(injector, handler.NewHealthCheckHandler)
	do.Provide(injector, handler.NewAuthHandler)
}
