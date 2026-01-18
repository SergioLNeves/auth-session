package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DBPath      string
	Environment string
	MaxConn     int
	MaxIdle     int
	MaxLifeTime time.Duration
}

type SQLiteStorage struct {
	db *gorm.DB
}

func NewSQLite(cfg *Config) (*SQLiteStorage, error) {
	// Criar diretório do banco de dados se não existir
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configurar nível de log
	logLevel := logger.Info
	if cfg.Environment == "production" {
		logLevel = logger.Warn
	}

	// Configuração do GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	}

	// Conectar ao banco de dados
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Obter SQL DB para configurações avançadas
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configurar pool de conexões
	sqlDB.SetMaxOpenConns(cfg.MaxConn)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifeTime)

	// Testar conexão
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to SQLite database at %s", cfg.DBPath)

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) GetDB() *gorm.DB {
	return s.db
}

func (s *SQLiteStorage) Ping() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}

func (s *SQLiteStorage) AutoMigrate(models ...any) error {
	if err := s.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	return nil
}

func InitDatabase(dbPath, environment string, maxConn, maxIdle int, maxLifeTime time.Duration) (*SQLiteStorage, error) {
	cfg := &Config{
		DBPath:      dbPath,
		Environment: environment,
		MaxConn:     maxConn,
		MaxIdle:     maxIdle,
		MaxLifeTime: maxLifeTime,
	}

	db, err := NewSQLite(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := db.AutoMigrate(GetModelsToMigrate()...); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
