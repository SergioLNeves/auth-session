package domain

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenClaims struct {
	UserID    string
	Email     string
	Name      string
	Avatar    string
	SessionID string
}

type TokenProvider interface {
	GenerateAccessToken(userID, email, name, avatar, sessionID string) (string, error)
	GenerateRefreshToken(userID, sessionID string) (string, error)
	ParseAccessToken(tokenString string) (*TokenClaims, error)
	ParseRefreshToken(tokenString string) (*TokenClaims, error)
}
