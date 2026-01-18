package config

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

type API struct {
	echo            *echo.Echo
	port            int
	shutdownTimeout time.Duration
}

func NewAPI(e *echo.Echo, port int, shutdownTimeout time.Duration) *API {
	return &API{
		echo:            e,
		port:            port,
		shutdownTimeout: shutdownTimeout,
	}
}

func (s *API) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		address := fmt.Sprintf(":%d", s.port)
		if err := s.echo.Start(address); err != nil {
			s.echo.Logger.Info("Shutting down the server")
		}
	}()

	<-quit
	s.echo.Logger.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		s.echo.Logger.Error("Server forced to shutdown, err:", err)
	}

	s.echo.Logger.Info("Server exited gracefully")
}
