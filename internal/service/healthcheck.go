package service

import (
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
)

const (
	WorkingStatus   = "WORKING"
	DatabaseHealthy = "healthy"
	DatabaseError   = "error"
)

type HealthCheckServiceImpl struct {
	db *storage.SQLiteStorage
}

func NewHealthCheckService(db *storage.SQLiteStorage) (domain.HealthCheckerService, error) {
	return &HealthCheckServiceImpl{db: db}, nil
}

func (h *HealthCheckServiceImpl) Check() (domain.HealthCheck, []error) {
	var errs []error

	dbStatus := DatabaseHealthy
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
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
