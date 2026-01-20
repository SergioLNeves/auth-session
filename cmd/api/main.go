package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/handler"
	validator "github.com/SergioLNeves/auth-session/internal/pkg"
	"github.com/SergioLNeves/auth-session/internal/repository"
	"github.com/SergioLNeves/auth-session/internal/service"
	"github.com/SergioLNeves/auth-session/internal/storage/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
)

var (
	injector *do.Injector
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Validator = validator.NewValidator()

	initDependencies()
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
		e.Logger.Fatal(fmt.Errorf("invoke healthcheck handler: %w", err))
	}

	e.GET("/health", healthCheckHandler.Check)
}

func configureAuthRoute(e *echo.Echo) {
	authHandler, err := do.Invoke[domain.AuthHandler](injector)
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("invoke auth handler: %w", err))
	}

	v1 := e.Group("/v1")
	authUser := v1.Group("/auth/user")
	authUser.POST("/create-account", authHandler.CreateAccount)
	authUser.POST("/login", authHandler.Login)
}

func initDependencies() {
	injector = do.New()

	do.Provide(injector, sqlite.NewSQLite)

	do.Provide(injector, repository.NewAuthRepository)

	do.Provide(injector, service.NewHealthCheckService)
	do.Provide(injector, service.NewAuthService)

	do.Provide(injector, handler.NewHealthCheckHandler)
	do.Provide(injector, handler.NewAuthHandler)
}
