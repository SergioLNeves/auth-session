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
	Active    bool      `gorm:"default:true"`
	Avatar    string
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

func GetModelsToMigrate() []any {
	return []any{
		&UserTable{},
		&SessionTable{},
	}
}
