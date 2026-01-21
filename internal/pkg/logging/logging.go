package logging

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/SergioLNeves/auth-session/internal/domain"
)

var logger *zap.Logger

func NewLogger(configs *domain.Config) *zap.Logger {
	logLevel := getLogLevel(configs.LogLevel)

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: getEncoding(configs),
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			MessageKey:     "messsage",
			CallerKey:      "caller",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	builtLogger, err := config.Build()
	if err != nil {
		log.Fatalf("it was not possible to initialiaze the log system: %s", err.Error())
	}
	logger = builtLogger
	return logger
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if logger == nil {
		panic("logger not initialized, call NewLogger first")
	}
	return logger
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// Debug logs a debug message with fields
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message with fields
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message with fields
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message with fields
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message with fields and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// getLogLevel returns the appropriate log level
func getLogLevel(loglevel string) zapcore.Level {
	switch loglevel {
	case "error":
		return zap.ErrorLevel
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	default:
		return zap.InfoLevel
	}
}

func getEncoding(configs *domain.Config) string {
	if configs.Env == "development" {
		return "console"
	}
	return "json"
}
