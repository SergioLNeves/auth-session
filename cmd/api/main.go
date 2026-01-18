package main

import (
	"github.com/SergioLNeves/auth-session/internal/handler"
	validator "github.com/SergioLNeves/auth-session/internal/pkg"
	"github.com/SergioLNeves/auth-session/internal/service"
	"github.com/SergioLNeves/auth-session/internal/storage"
	"github.com/gookit/slog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db, err := storage.InitDatabase()
	if err != nil {
		slog.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Validator = validator.NewValidator()

	configureHealthcheckRoute(e, db)

	StartServer(e)
}

func StartServer(e *echo.Echo) {
	slog.Println("Starting server on :8080")
	if err := e.Start(":8080"); err != nil {
		slog.Fatalf("Server failed: %v", err)
	}
}

func configureHealthcheckRoute(e *echo.Echo, db *storage.SQLiteStorage) {
	healthService, _ := service.NewHealthCheckService(db)
	healthCheckHandler, _ := handler.NewHealthCheckHandler(healthService)

	e.GET("/health", healthCheckHandler.Check)
}
