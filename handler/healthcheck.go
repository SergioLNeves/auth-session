package handler

import (
	"fmt"
	"net/http"

	"github.com/SergioLNeves/auth-session/domain"
	"github.com/labstack/echo/v4"
)

type HealthCheckHandlerImpl struct {
	healthCheckService domain.HealthCheckerService
}

func NewHealthCheckHandler(healthCheckService domain.HealthCheckerService) (domain.HealthCheckHandler, error) {
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

	return ctx.JSON(http.StatusOK, check.Status)
}
