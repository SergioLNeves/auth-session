package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
)

var TableSession = "session"

type SessionRepositoryImpl struct {
	db storage.Storage
}

func NewSessionRepository(i *do.Injector) (domain.SessionRepository, error) {
	db := do.MustInvoke[storage.Storage](i)
	return &SessionRepositoryImpl{db: db}, nil
}

func (r *SessionRepositoryImpl) CreateSession(ctx context.Context, session *domain.Session) error {
	if err := r.db.Insert(ctx, TableSession, session); err != nil {
		return err
	}
	return nil
}

func (r *SessionRepositoryImpl) FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.FindByID(ctx, TableSession, sessionID, &session); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, err
	}

	if !session.ExpiresAt.IsZero() && session.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrSessionNotFound
	}

	return &session, nil
}

func (r *SessionRepositoryImpl) DeleteSession(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.FindOneAndDelete(ctx, TableSession, sessionID, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepositoryImpl) UpdateSessionExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := db.WithContext(ctx).Table(TableSession).Where("id = ?", sessionID).Update("expires_at", expiresAt)
	if result.Error != nil {
		return fmt.Errorf("failed to update session expiry: %w", result.Error)
	}

	return nil
}

func (r *SessionRepositoryImpl) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := db.WithContext(ctx).Table(TableSession).Where("user_id = ?", userID).Delete(&domain.Session{})
	return result.Error
}

func (r *SessionRepositoryImpl) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	db, ok := r.db.GetDB().(*gorm.DB)
	if !ok {
		return 0, fmt.Errorf("failed to get database instance")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result := db.WithContext(ctx).Table(TableSession).Where("expires_at <= ?", time.Now()).Delete(&domain.Session{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", result.Error)
	}

	return result.RowsAffected, nil
}
