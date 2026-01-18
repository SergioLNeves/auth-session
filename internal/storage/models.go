package storage

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTable struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	Name      string    `gorm:"not null"`
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func GetModelsToMigrate() []any {
	return []any{
		&UserTable{},
	}
}
