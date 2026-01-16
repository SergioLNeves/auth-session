package main

import (
	"github.com/SergioLNeves/auth-session/handler"
	validator "github.com/SergioLNeves/auth-session/internal/pkg"
	"github.com/SergioLNeves/auth-session/service"
	"github.com/gookit/slog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Validator = validator.NewValidator()

	configureHealthcheckRoute(e)

	StartServer(e)
}

func StartServer(e *echo.Echo) {
	slog.Println("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		slog.Fatalf("Server failed: %v", err)
	}
}

func configureLogger() {
	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})
}

func configureHealthcheckRoute(e *echo.Echo) {
	healthService, _ := service.NewHealthCheckService()
	healthCheckHandler, _ := handler.NewHealthCheckHandler(healthService)

	e.GET("/health", healthCheckHandler.Check)
}
