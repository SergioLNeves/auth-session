package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SergioLNeves/auth-session/internal/config"
	"github.com/SergioLNeves/auth-session/internal/storage"
	"github.com/samber/do"
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

func NewSQLite(i *do.Injector) (storage.Storage, error) {
	return newSQLite(&Config{
		DBPath:      config.Env.SQL.DBPath,
		Environment: config.Env.Env,
		MaxConn:     config.Env.SQL.MaxConn,
		MaxIdle:     config.Env.SQL.MaxIdle,
		MaxLifeTime: config.Env.SQL.MaxLifeTime,
	})
}

func newSQLite(cfg *Config) (storage.Storage, error) {
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConn)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifeTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqliteDB := &SQLiteStorage{db: db}

	if err := sqliteDB.AutoMigrate(GetModelsToMigrate()...); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return sqliteDB, nil
}

func (s *SQLiteStorage) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
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

	return nil
}

func (s *SQLiteStorage) AutoMigrate(models ...any) error {
	if err := s.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) Insert(ctx context.Context, table string, data any) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := s.db.WithContext(ctx).Table(table).Create(data)
	if result.Error != nil {
		return fmt.Errorf("failed to insert data: %w", result.Error)
	}
	return nil
}

func (s *SQLiteStorage) Update(ctx context.Context, table string, data any) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := s.db.WithContext(ctx).Table(table).Save(data)
	if result.Error != nil {
		return fmt.Errorf("failed to update data: %w", result.Error)
	}
	return nil
}

func (s *SQLiteStorage) FindOneAndDelete(ctx context.Context, table string, id any, dest any) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := s.db.WithContext(ctx).Table(table).Where("id = ?", id).First(dest)
	if result.Error != nil {
		return result.Error
	}

	if err := s.db.WithContext(ctx).Table(table).Where("id = ?", id).Delete(dest).Error; err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) FindByID(ctx context.Context, table string, id any, dest any) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := s.db.WithContext(ctx).Table(table).Where("id = ?", id).First(dest)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SQLiteStorage) GetDB() any {
	return s.db
}

func (s *SQLiteStorage) FindByEmail(ctx context.Context, table, email string, dest any) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result := s.db.WithContext(ctx).Table(table).Where("email = ?", email).First(dest)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
