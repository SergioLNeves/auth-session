package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	ErrEmailAlreadyExists = fmt.Errorf("Error Email Already Exists")
)

type CreateAccountRequest struct {
	Name     string `form:"name" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
}

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

type AuthHandler interface {
	CreateAccount(c echo.Context) error
	Login(c echo.Context) error
}

type AuthService interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*User, error)
}

type AuthRepository interface {
	CreateUser(ctx context.Context, user *User) error
}
