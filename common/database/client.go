// Package database provides MySQL database connection management
// optimized for high-concurrency microservices in the Aether Defense System.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// Config holds database connection configuration.
type Config struct {
	// Note: We mark fields as optional for go-zero conf.MustLoad.
	// Defaults are applied in DefaultConfig()/NewClient when values are zero.
	//lint:ignore SA5008 go-zero config uses json tag options like ",optional"; not for encoding/json.
	DSN string `json:"dsn,optional"` // Data Source Name

	//lint:ignore SA5008 go-zero config uses json tag options like ",optional"; not for encoding/json.
	MaxOpenConns int `json:"max_open_conns,optional"`

	//lint:ignore SA5008 go-zero config uses json tag options like ",optional"; not for encoding/json.
	MaxIdleConns int `json:"max_idle_conns,optional"`

	//lint:ignore SA5008 go-zero config uses json tag options like ",optional"; not for encoding/json.
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime,optional"`

	//lint:ignore SA5008 go-zero config uses json tag options like ",optional"; not for encoding/json.
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time,optional"`
}

// DefaultConfig returns a default database configuration optimized for high concurrency.
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConns:    100,              // High connection pool for concurrent operations
		MaxIdleConns:    10,               // Keep connections warm
		ConnMaxLifetime: 30 * time.Minute, // Recycle connections periodically
		ConnMaxIdleTime: 10 * time.Minute, // Close idle connections
	}
}

// Client wraps the database connection with connection pool management.
type Client struct {
	db *sql.DB
}

// NewClient creates a new database client with the given configuration.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("database config is required")
	}

	if config.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	// Use default values if not set
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = DefaultConfig().MaxOpenConns
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = DefaultConfig().MaxIdleConns
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = DefaultConfig().ConnMaxLifetime
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = DefaultConfig().ConnMaxIdleTime
	}

	db, err := sql.Open("mysql", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			// Log close error but don't override ping error
			_ = closeErr
		}
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{db: db}, nil
}

// DB returns the underlying *sql.DB instance.
func (c *Client) DB() *sql.DB {
	return c.db
}

// Close closes the database connection.
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive.
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// BuildDSN builds a MySQL DSN string from individual components.
func BuildDSN(user, password, host, port, dbName string, params map[string]string) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	if len(params) > 0 {
		dsn += "?"
		first := true
		for k, v := range params {
			if !first {
				dsn += "&"
			}
			dsn += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	} else {
		// Default parameters for production use
		dsn += "?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=3s&writeTimeout=3s"
	}

	return dsn
}
