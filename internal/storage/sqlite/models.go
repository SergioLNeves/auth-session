package sqlite

import (
	"time"

	"github.com/google/uuid"
)

type UserTable struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	Avatar    string
	DeletedAt *time.Time `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (UserTable) TableName() string { return "user" }

type SessionTable struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SessionTable) TableName() string { return "session" }

type DeviceTable struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	UserAgent string    `gorm:"not null;default:'unknown'"`
	IPAddress string    `gorm:"not null;default:'unknown'"`
	CreatedAt time.Time
}

func (DeviceTable) TableName() string { return "device" }

func GetModelsToMigrate() []any {
	return []any{
		&UserTable{},
		&SessionTable{},
		&DeviceTable{},
	}
}
