package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrSessionNotFound = fmt.Errorf("session not found")

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time `gorm:"not null;index"`
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session) error
	FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	UpdateSessionExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error
}
