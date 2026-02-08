package sqlite

import (
	"time"

	"github.com/google/uuid"
)

type UserTable struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SessionTable struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetModelsToMigrate() []any {
	return []any{
		&UserTable{},
		&SessionTable{},
	}
}
