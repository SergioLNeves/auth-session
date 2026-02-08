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

func GetModelsToMigrate() []any {
	return []any{
		&UserTable{},
	}
}
