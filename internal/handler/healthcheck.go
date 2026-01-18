package handler

import (
	"fmt"
	"net/http"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

type HealthCheckHandlerImpl struct {
	healthCheckService domain.HealthCheckerService
}

func NewHealthCheckHandler(i *do.Injector) (domain.HealthCheckHandler, error) {
	healthCheckService := do.MustInvoke[domain.HealthCheckerService](i)
	if healthCheckService == nil {
		return nil, fmt.Errorf("failed to initialize health check service dependency")
	}

	return &HealthCheckHandlerImpl{
		healthCheckService: healthCheckService,
	}, nil
}

func (h HealthCheckHandlerImpl) Check(ctx echo.Context) error {
	check, err := h.healthCheckService.Check()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusOK, check)
}
