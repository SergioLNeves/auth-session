package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	Cost = 12
)

func HashPassword(password string) (*string, error) {
	passwordByte := []byte(password)

	hashedBytes, err := bcrypt.GenerateFromPassword(passwordByte, Cost)
	if err != nil {
		return nil, fmt.Errorf("Error during generated hash: %w", err)
	}

	passwordHash := string(hashedBytes)

	return &passwordHash, nil
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}

// NeedRehash verifica se hash precisa ser atualizado
func NeedRehash(hash string, cost int) (bool, error) {
	// Extrai o cost atual do hash armazenado
	hashCost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return false, err
	}
	// Se o cost do hash é menor que o cost atual, precisa atualizar
	return hashCost < cost, nil
}
