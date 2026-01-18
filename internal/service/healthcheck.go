package service

import (
	"context"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
	"github.com/samber/do"
)

const (
	WorkingStatus   = "WORKING"
	DatabaseHealthy = "healthy"
	DatabaseError   = "error"
)

type HealthCheckServiceImpl struct {
	db storage.Storage
}

func NewHealthCheckService(i *do.Injector) (domain.HealthCheckerService, error) {
	db := do.MustInvoke[storage.Storage](i)
	return &HealthCheckServiceImpl{db: db}, nil
}

func (h *HealthCheckServiceImpl) Check() (domain.HealthCheck, []error) {
	var errs []error

	dbStatus := DatabaseHealthy
	if h.db != nil {
		if err := h.db.Ping(context.Background()); err != nil {
			dbStatus = DatabaseError
			errs = append(errs, err)
		}
	}

	healthCheck := domain.HealthCheck{
		Status:   WorkingStatus,
		Database: dbStatus,
	}

	return healthCheck, errs
}
