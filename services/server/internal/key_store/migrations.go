package key_store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
)

const currentMigrationVersion int64 = 1

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

// Migrate applies pending key-store migrations.
func (s *Store) Migrate(ctx context.Context) error {
	provider, err := newMigrationProvider(s.db)
	if err != nil {
		return fmt.Errorf("create key store migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply key store migrations: %w", err)
	}
	return nil
}

func newMigrationProvider(db *sql.DB) (*goose.Provider, error) {
	migrationFS, err := fs.Sub(embeddedMigrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("open embedded key store migrations: %w", err)
	}
	return goose.NewProvider(
		goose.DialectPostgres,
		db,
		migrationFS,
		goose.WithDisableGlobalRegistry(true),
	)
}
