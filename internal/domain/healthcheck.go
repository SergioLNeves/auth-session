package domain

import "github.com/labstack/echo/v4"

type HealthCheck struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

type HealthCheckerService interface {
	Check() (HealthCheck, []error)
}

type HealthCheckHandler interface {
	Check(ctx echo.Context) error
}
