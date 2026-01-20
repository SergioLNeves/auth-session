package repository

import (
	"context"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
	"github.com/samber/do"
)

type AuthRepositoryImpl struct {
	db storage.Storage
}

func NewAuthRepository(i *do.Injector) (domain.AuthRepository, error) {
	db := do.MustInvoke[storage.Storage](i)
	return &AuthRepositoryImpl{db: db}, nil
}

func (r *AuthRepositoryImpl) CreateUser(ctx context.Context, user *domain.User) error {
	if err := r.db.Insert(ctx, user); err != nil {
		return err
	}
	return nil
}
