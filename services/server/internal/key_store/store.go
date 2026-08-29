package key_store

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config describes the dedicated key metadata database connection.
type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// Store owns the key metadata database connection pool.
type Store struct {
	db *sql.DB
}

// Open creates and verifies a key-store database connection pool.
func Open(ctx context.Context, config Config) (*Store, error) {
	connectionURL, err := config.connectionURL()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("open key store: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping key store: %w", err)
	}

	return &Store{db: db}, nil
}

// DB returns the underlying pooled database handle.
func (s *Store) DB() *sql.DB {
	return s.db
}

// Close closes the key-store database connection pool.
func (s *Store) Close() error {
	return s.db.Close()
}

func (config Config) connectionURL() (string, error) {
	if config.Host == "" || config.Database == "" || config.User == "" {
		return "", fmt.Errorf("key store host, database, and user are required")
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}

	connectionURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.User, config.Password),
		Host:   net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
		Path:   "/" + config.Database,
	}
	query := connectionURL.Query()
	query.Set("sslmode", config.SSLMode)
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String(), nil
}
