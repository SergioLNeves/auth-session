package domain

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenProvider interface {
	GenerateAccessToken(userID string, email string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
}
