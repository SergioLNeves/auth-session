package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
)

var TableSession = "session_tables"

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
			return nil, nil
		}
		return nil, err
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
