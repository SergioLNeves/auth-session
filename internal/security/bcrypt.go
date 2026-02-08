package security

import (
	"fmt"

	"github.com/samber/do"
	"golang.org/x/crypto/bcrypt"

	"github.com/SergioLNeves/auth-session/internal/domain"
)

const (
	Cost = 12
)

type BcryptHasher struct{}

func NewBcryptHasher(_ *do.Injector) (domain.PasswordHasher, error) {
	return &BcryptHasher{}, nil
}

func (b *BcryptHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

func (b *BcryptHasher) Check(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
