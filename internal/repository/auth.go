package repository

import (
	"context"

	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/SergioLNeves/auth-session/internal/storage"
	"github.com/samber/do"
)

type authRepository struct {
	db storage.Storage
}

func NewAuthRepository(i *do.Injector) (domain.AuthRepository, error) {
	s := do.MustInvoke[storage.Storage](i)
	return &authRepository{db: s}, nil
}

// GetUserByEmail retrieves a user from the database by their email address.
func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.FindByEmail(ctx, email, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser saves a new user to the database.
func (r *authRepository) CreateUser(ctx context.Context, user *domain.User) error {
	if err := r.db.Insert(ctx, user); err != nil {
		return err
	}
	return nil
}
