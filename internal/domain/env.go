package domain

import "time"

type Config struct {
	Env      string `env:"ENV,default=development"`
	Port     int    `env:"PORT,default=8080"`
	LogLevel string `env:"LOG_LEVEL,default:debug"`
	Keys     KeysConfig
	Token    TokenConfig
	SQL      SQLConfig
}

type KeysConfig struct {
	PrivateKeyPath string `env:"PRIVATE_KEY_PATH,required=true"`
	PublicKeyPath  string `env:"PUBLIC_KEY_PATH,required=true"`
}

type TokenConfig struct {
	AccessTokenExpiry  int `env:"ACCESS_TOKEN_EXPIRY,default=60"`
	RefreshTokenExpiry int `env:"REFRESH_TOKEN_EXPIRY,default=10080"`
}

type SQLConfig struct {
	DBPath      string        `env:"DB_PATH,default=./data/auth-session.db"`
	MaxConn     int           `env:"DB_MAX_CONN,default=10"`
	MaxIdle     int           `env:"DB_MAX_IDLE,default=5"`
	MaxLifeTime time.Duration `env:"DB_MAX_LIFETIME,default=1h"`
}
