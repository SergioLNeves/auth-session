package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateAccountRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthHandler interface {
	CreateAccount(c echo.Context) error
	Login(c echo.Context) error
}

type AuthService interface {
	Authenticate(ctx context.Context, email, password string) (*User, error)
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*User, error)
}

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
}
