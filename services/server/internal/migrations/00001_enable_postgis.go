package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upEnablePostGIS, downEnablePostGIS)
}

func upEnablePostGIS(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS postgis")
	return err
}

func downEnablePostGIS(context.Context, *sql.Tx) error {
	// The database image owns PostGIS, so rollback intentionally retains it.
	return nil
}
