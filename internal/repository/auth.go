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

var (
	TableUser = "user"
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
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepositoryImpl) FindUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	db := r.db.GetDB().(*gorm.DB)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	var user domain.User
	if err := db.WithContext(ctx).Table(TableUser).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepositoryImpl) UpdateUser(ctx context.Context, user *domain.User) error {
	return r.db.Update(ctx, TableUser, user)
}

func (r *AuthRepositoryImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	db := r.db.GetDB().(*gorm.DB)
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	now := time.Now()
	result := db.WithContext(ctx).Table(TableUser).Where("id = ?", id).Update("deleted_at", now)
	return result.Error
}

func (r *AuthRepositoryImpl) DeleteDeactivatedUsers(ctx context.Context) (int64, error) {
	db := r.db.GetDB().(*gorm.DB)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	result := db.WithContext(ctx).Table(TableUser).
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", sevenDaysAgo).
		Delete(&domain.User{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete deactivated users: %w", result.Error)
	}
	return result.RowsAffected, nil
}
