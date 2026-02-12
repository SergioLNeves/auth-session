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
	publicKey          *rsa.PublicKey
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewJWTProvider(_ *do.Injector) (domain.TokenProvider, error) {
	privData, err := os.ReadFile(config.Env.Keys.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	pubData, err := os.ReadFile(config.Env.Keys.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &JWTProvider{
		privateKey:         privateKey,
		publicKey:          publicKey,
		accessTokenExpiry:  time.Duration(config.Env.Token.AccessTokenExpiry) * time.Minute,
		refreshTokenExpiry: time.Duration(config.Env.Token.RefreshTokenExpiry) * time.Minute,
	}, nil
}

func (j *JWTProvider) GenerateAccessToken(sessionID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": sessionID,
		"iat": now.Unix(),
		"exp": now.Add(j.accessTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return signed, nil
}

func (j *JWTProvider) GenerateRefreshToken(userID string, sessionID string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":        userID,
		"session_id": sessionID,
		"iat":        now.Unix(),
		"exp":        now.Add(j.refreshTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signed, nil
}

func (j *JWTProvider) ParseAccessToken(tokenString string) (*domain.AccessTokenClaims, error) {
	token, err := j.parseToken(tokenString, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &domain.AccessTokenClaims{
		SessionID: claims["sub"].(string),
	}, nil
}

func (j *JWTProvider) ParseRefreshToken(tokenString string) (*domain.RefreshTokenClaims, error) {
	token, err := j.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &domain.RefreshTokenClaims{
		UserID:    claims["sub"].(string),
		SessionID: claims["session_id"].(string),
	}, nil
}

func (j *JWTProvider) parseToken(tokenString string, opts ...jwt.ParserOption) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	}

	token, err := jwt.Parse(tokenString, keyFunc, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token, nil
}
