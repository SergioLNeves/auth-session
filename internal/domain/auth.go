package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	ErrEmailAlreadyExists     = fmt.Errorf("Error Email Already Exists")
	ErrInvalidCredentials     = fmt.Errorf("Error Invalid Credentials")
	ErrUserNotFound           = fmt.Errorf("Error User Not Found")
	ErrInvalidCurrentPassword = fmt.Errorf("Error Invalid Current Password")
	ErrUserDeactivated        = fmt.Errorf("Error User Deactivated")
	ErrUserNotDeactivated     = fmt.Errorf("Error User Not Deactivated")
)

type CreateAccountRequest struct {
	Name     string `form:"name" validate:"required,name"`
	Avatar   string
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
}

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name      string
	Email     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	Avatar    string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginRequest struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `form:"password" validate:"required"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `form:"current_password" validate:"required"`
	NewPassword     string `form:"new_password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	Name   string `form:"name" validate:"omitempty,name"`
	Email  string `form:"email" validate:"omitempty,email"`
	Avatar string `form:"avatar"`
}

type UserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

type AuthHandler interface {
	CreateAccount(c echo.Context) error
	Login(c echo.Context) error
	Logout(c echo.Context) error
	UpdatePassword(c echo.Context) error
	UpdateUser(c echo.Context) error
	Me(c echo.Context) error
	DeleteUser(c echo.Context) error
	ReactivateAccount(c echo.Context) error
}

type AuthService interface {
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*AuthResponse, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Logout(ctx context.Context, sessionID string) error
	UpdatePassword(ctx context.Context, userID string, req UpdatePasswordRequest) error
	UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, userID string) error
	ReactivateAccount(ctx context.Context, req LoginRequest) (*AuthResponse, error)
}

type AuthRepository interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DeleteDeactivatedUsers(ctx context.Context) (int64, error)
}
