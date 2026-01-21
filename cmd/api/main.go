package main

import (
	"time"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/handler"
	"github.com/SergioLNeves/auth-session/internal/pkg/logging"
	validator "github.com/SergioLNeves/auth-session/internal/pkg/validator"
	"github.com/SergioLNeves/auth-session/internal/repository"
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
	authHandler, err := do.Invoke[domain.AuthHandler](injector)
	if err != nil {
		logger.Fatal("invoke auth handler", zap.Error(err))
	}

	v1 := e.Group("/v1")
	authUser := v1.Group("/auth/user")
	authUser.POST("/create-account", authHandler.CreateAccount)
	authUser.POST("/login", authHandler.Login)
}

func initDependencies(logger *zap.Logger) {
	injector = do.New()

	do.ProvideValue(injector, logger)

	do.Provide(injector, sqlite.NewSQLite)

	do.Provide(injector, repository.NewAuthRepository)

	do.Provide(injector, service.NewHealthCheckService)
	do.Provide(injector, service.NewAuthService)

	do.Provide(injector, handler.NewHealthCheckHandler)
	do.Provide(injector, handler.NewAuthHandler)
}
