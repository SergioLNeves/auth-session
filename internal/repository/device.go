package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
)

var TableDevice = "device"

type DeviceRepositoryImpl struct {
	db storage.Storage
}

func NewDeviceRepository(i *do.Injector) (domain.DeviceRepository, error) {
	db := do.MustInvoke[storage.Storage](i)
	return &DeviceRepositoryImpl{db: db}, nil
}

func (r *DeviceRepositoryImpl) CreateDevice(ctx context.Context, device *domain.Device) error {
	return r.db.Insert(ctx, TableDevice, device)
}

func (r *DeviceRepositoryImpl) FindDeviceBySessionID(ctx context.Context, sessionID uuid.UUID) (*domain.Device, error) {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var device domain.Device
	result := db.WithContext(ctx).Table(TableDevice).Where("session_id = ?", sessionID).First(&device)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find device by session ID: %w", result.Error)
	}

	return &device, nil
}

func (r *DeviceRepositoryImpl) DeleteDeviceBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := db.WithContext(ctx).Table(TableDevice).Where("session_id = ?", sessionID).Delete(&domain.Device{})
	return result.Error
}

func (r *DeviceRepositoryImpl) DeleteDevicesByUserID(ctx context.Context, userID uuid.UUID) error {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := db.WithContext(ctx).Table(TableDevice).
		Where("session_id IN (SELECT id FROM session WHERE user_id = ?)", userID).
		Delete(&domain.Device{})
	return result.Error
}

func (r *DeviceRepositoryImpl) DeleteExpiredDevices(ctx context.Context) (int64, error) {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return 0, fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := db.WithContext(ctx).Table(TableDevice).
		Where("session_id NOT IN (SELECT id FROM session)").
		Delete(&domain.Device{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired devices: %w", result.Error)
	}

	return result.RowsAffected, nil
}
