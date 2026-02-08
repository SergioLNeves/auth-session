package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session) error
	FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	DeactivateSession(ctx context.Context, sessionID uuid.UUID) error
}
