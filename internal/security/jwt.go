package security

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/do"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/domain"
)

type JWTProvider struct {
	privateKey         *rsa.PrivateKey
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewJWTProvider(_ *do.Injector) (domain.TokenProvider, error) {
	keyData, err := os.ReadFile(config.Env.Keys.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &JWTProvider{
		privateKey:         privateKey,
		accessTokenExpiry:  time.Duration(config.Env.Token.AccessTokenExpiry) * time.Minute,
		refreshTokenExpiry: time.Duration(config.Env.Token.RefreshTokenExpiry) * time.Minute,
	}, nil
}

func (j *JWTProvider) GenerateAccessToken(userID string, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(j.accessTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signed, nil
}

func (j *JWTProvider) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(j.refreshTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signed, nil
}
