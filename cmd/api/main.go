package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/handler"
	validator "github.com/SergioLNeves/auth-session/internal/pkg"
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

	api := config.NewAPI(e, config.Env.Port, 10*time.Second)
	api.Start()
}

func configureHealthcheckRoute(e *echo.Echo) {
	healthCheckHandler, err := do.Invoke[domain.HealthCheckHandler](injector)
	if err != nil {
		e.Logger.Fatal(fmt.Errorf("Invoke healthcheck handle: %w", err))
	}

	e.GET("/health", healthCheckHandler.Check)
}

func initDependencies() {
	injector = do.New()

	do.Provide(injector, sqlite.NewSQLite)
	do.Provide(injector, service.NewHealthCheckService)
	do.Provide(injector, handler.NewHealthCheckHandler)
}
