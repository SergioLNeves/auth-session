package domain

import "github.com/google/uuid"

type UserRequest struct {
	SessionID uuid.UUID
	Login     string
	Password  string
}

type LoginHandler interface {
	Login() error
}

type LoginService interface {
}

type LoginRepository interface {
}
