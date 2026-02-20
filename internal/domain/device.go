package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const UnknownDevice = "unknown"

type DeviceInfo struct {
	UserAgent string
	IPAddress string
}

func NewDeviceInfo(userAgent, ip string) DeviceInfo {
	if userAgent == "" {
		userAgent = UnknownDevice
	}
	if ip == "" {
		ip = UnknownDevice
	}
	return DeviceInfo{UserAgent: userAgent, IPAddress: ip}
}

type Device struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	UserAgent string    `gorm:"not null;default:'unknown'"`
	IPAddress string    `gorm:"not null;default:'unknown'"`
	CreatedAt time.Time
}

type DeviceRepository interface {
	CreateDevice(ctx context.Context, device *Device) error
	FindDeviceBySessionID(ctx context.Context, sessionID uuid.UUID) (*Device, error)
	DeleteDeviceBySessionID(ctx context.Context, sessionID uuid.UUID) error
	DeleteDevicesByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredDevices(ctx context.Context) (int64, error)
}
