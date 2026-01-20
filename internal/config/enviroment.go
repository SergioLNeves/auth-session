package config

import (
	"fmt"

	"github.com/Netflix/go-env"
	"github.com/SergioLNeves/auth-session/internal/domain"
	"github.com/joho/godotenv"
)

var Env domain.Config

func LoadEnv() error {
	_ = godotenv.Load() //nolint:errcheck // .env file is optional

	if _, err := env.UnmarshalFromEnviron(&Env); err != nil {
		return fmt.Errorf("init env: %w", err)
	}
	return nil
}
