package repository

import (
	"context"
	"errors"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
	"github.com/samber/do"
	"gorm.io/gorm"
)

var (
	TableUser = "user_tables"
)

type AuthRepositoryImpl struct {
	db storage.Storage
}

func NewAuthRepository(i *do.Injector) (domain.AuthRepository, error) {
	db := do.MustInvoke[storage.Storage](i)
	return &AuthRepositoryImpl{db: db}, nil
}

func (r *AuthRepositoryImpl) CreateUser(ctx context.Context, user *domain.User) error {
	if err := r.db.Insert(ctx, TableUser, user); err != nil {
		return err
	}
	return nil
}

func (r *AuthRepositoryImpl) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.FindByEmail(ctx, TableUser, email, &user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
