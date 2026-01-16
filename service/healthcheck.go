package service

import (
	"github.com/SergioLNeves/auth-session/domain"
)

const (
	WorkingStatus = "WORKING"
)

type HealthCheckServiceImpl struct {
}

func NewHealthCheckService() (domain.HealthCheckerService, error) {
	return HealthCheckServiceImpl{}, nil
}

func (h HealthCheckServiceImpl) Check() (domain.HealthCheck, []error) {
	healthCheck := domain.HealthCheck{Status: WorkingStatus}

	var errs []error

	return healthCheck, errs
}
