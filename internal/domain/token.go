package domain

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AccessTokenClaims struct {
	SessionID string
}

type RefreshTokenClaims struct {
	UserID    string
	SessionID string
}

type TokenProvider interface {
	GenerateAccessToken(sessionID string) (string, error)
	GenerateRefreshToken(userID, sessionID string) (string, error)
	ParseAccessToken(tokenString string) (*AccessTokenClaims, error)
	ParseRefreshToken(tokenString string) (*RefreshTokenClaims, error)
}
