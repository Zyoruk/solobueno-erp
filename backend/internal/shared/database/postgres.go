// Package database provides database connection utilities for the application.
package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database connection configuration.
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// DefaultConfig returns configuration from environment variables with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "solobueno"),
		Password:        getEnv("DB_PASSWORD", "solobueno"),
		Database:        getEnv("DB_NAME", "solobueno_dev"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}
}

// NewConnection creates a new GORM database connection.
func NewConnection(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	// Configure GORM logger based on environment
	logLevel := logger.Silent
	if getEnv("DB_LOG_LEVEL", "silent") == "info" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		PrepareStmt:            true, // Prepared statement cache for performance
		SkipDefaultTransaction: true, // Skip default transaction for better performance
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// MustConnect creates a connection and panics if it fails.
// Use this for application startup where failure is fatal.
func MustConnect(cfg Config) *gorm.DB {
	db, err := NewConnection(cfg)
	if err != nil {
		panic(err)
	}
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
