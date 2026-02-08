package domain

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenClaims struct {
	UserID    string
	Email     string
	SessionID string
}

type TokenProvider interface {
	GenerateAccessToken(userID string, email string, sessionID string) (string, error)
	GenerateRefreshToken(userID string, sessionID string) (string, error)
	ParseAccessToken(tokenString string) (*TokenClaims, error)
}
