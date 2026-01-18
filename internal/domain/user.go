package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Name      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
