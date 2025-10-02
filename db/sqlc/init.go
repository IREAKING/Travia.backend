package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"travia.backend/config"
)

// InitDB initializes and returns a database connection pool
func InitDB(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	// sslmode is configurable via env; default to require for Cloud environments
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "require"
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		sslMode,
	)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool configuration optimized for Supabase
	poolConfig.MaxConns = 5                                                      // Reduced for Supabase connection limits
	poolConfig.MinConns = 1                                                      // Minimum number of connections
	poolConfig.MaxConnLifetime = 30 * time.Minute                                // Shorter lifetime for Supabase
	poolConfig.MaxConnIdleTime = 5 * time.Minute                                 // Shorter idle time for Supabase
	poolConfig.HealthCheckPeriod = 30 * time.Second                              // More frequent health checks
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol // Simple protocol for Supabase
	poolConfig.ConnConfig.StatementCacheCapacity = 0                             // Disable statement cache
	poolConfig.ConnConfig.DescriptionCacheCapacity = 0                           // Disable description cache

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool initialized successfully")
	return pool, nil
}

// CloseDB closes the database connection pool
func CloseDB(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
		log.Println("Database connection pool closed")
	}
}
