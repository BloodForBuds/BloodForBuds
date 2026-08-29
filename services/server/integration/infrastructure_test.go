//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/BloodForBuds/BloodForBuds/services/server/internal/app_store"
	"github.com/BloodForBuds/BloodForBuds/services/server/internal/key_store"
	"github.com/BloodForBuds/BloodForBuds/services/server/internal/kms"
	"github.com/jackc/pgx/v5"
)

const expectedMigrationVersion int64 = 1

func TestInfrastructure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	t.Run("app store", func(t *testing.T) {
		config := appStoreConfig(t)
		adminStore, err := app_store.Open(ctx, config)
		if err != nil {
			t.Fatalf("open app store admin connection: %v", err)
		}
		t.Cleanup(func() {
			if err := adminStore.Close(); err != nil {
				t.Errorf("close app store admin connection: %v", err)
			}
		})

		config.Database = createTestDatabase(t, ctx, adminStore.DB(), "app")
		store, err := app_store.Open(ctx, config)
		if err != nil {
			t.Fatalf("open test app store: %v", err)
		}
		t.Cleanup(func() {
			if err := store.Close(); err != nil {
				t.Errorf("close app store: %v", err)
			}
		})

		if err := store.Migrate(ctx); err != nil {
			t.Fatalf("migrate app store: %v", err)
		}

		var postGISVersion string
		if err := store.DB().QueryRowContext(
			ctx,
			"SELECT extversion FROM pg_extension WHERE extname = 'postgis'",
		).Scan(&postGISVersion); err != nil {
			t.Fatalf("query PostGIS version: %v", err)
		}
		if postGISVersion == "" {
			t.Fatal("expected PostGIS to be enabled")
		}

		assertMigrationVersion(t, ctx, store.DB(), expectedMigrationVersion)
	})

	t.Run("key store", func(t *testing.T) {
		config := keyStoreConfig(t)
		adminStore, err := key_store.Open(ctx, config)
		if err != nil {
			t.Fatalf("open key store admin connection: %v", err)
		}
		t.Cleanup(func() {
			if err := adminStore.Close(); err != nil {
				t.Errorf("close key store admin connection: %v", err)
			}
		})

		config.Database = createTestDatabase(t, ctx, adminStore.DB(), "key")
		store, err := key_store.Open(ctx, config)
		if err != nil {
			t.Fatalf("open test key store: %v", err)
		}
		t.Cleanup(func() {
			if err := store.Close(); err != nil {
				t.Errorf("close key store: %v", err)
			}
		})

		if err := store.Migrate(ctx); err != nil {
			t.Fatalf("migrate key store: %v", err)
		}

		var schemaName string
		if err := store.DB().QueryRowContext(
			ctx,
			"SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'key_store'",
		).Scan(&schemaName); err != nil {
			t.Fatalf("query key-store schema: %v", err)
		}
		if schemaName != "key_store" {
			t.Fatalf("expected key_store schema, got %q", schemaName)
		}

		assertMigrationVersion(t, ctx, store.DB(), expectedMigrationVersion)
	})

	t.Run("OpenBao", func(t *testing.T) {
		client, err := kms.New(kms.Config{
			Address: requiredEnvironment(t, "BAO_ADDR"),
			Token:   os.Getenv("BAO_TOKEN"),
		})
		if err != nil {
			t.Fatalf("create OpenBao client: %v", err)
		}

		health, err := client.Health(ctx)
		if err != nil {
			t.Fatalf("check OpenBao health: %v", err)
		}
		if health.Version == "" {
			t.Fatal("expected OpenBao to report a version")
		}
		if !health.Initialized || health.Sealed {
			t.Fatalf(
				"expected development OpenBao to be initialized and unsealed, got initialized=%t sealed=%t",
				health.Initialized,
				health.Sealed,
			)
		}
	})
}

func createTestDatabase(t *testing.T, ctx context.Context, db *sql.DB, storeName string) string {
	t.Helper()

	databaseName := fmt.Sprintf("bloodforbuds_%s_integration_%d", storeName, time.Now().UnixNano())
	identifier := pgx.Identifier{databaseName}.Sanitize()
	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+identifier); err != nil {
		t.Fatalf("create integration database %s: %v", databaseName, err)
	}

	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+identifier+" WITH (FORCE)"); err != nil {
			t.Errorf("drop integration database %s: %v", databaseName, err)
		}
	})
	return databaseName
}

func assertMigrationVersion(t *testing.T, ctx context.Context, db *sql.DB, expected int64) {
	t.Helper()

	var version int64
	if err := db.QueryRowContext(
		ctx,
		"SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = true",
	).Scan(&version); err != nil {
		t.Fatalf("query Goose version: %v", err)
	}
	if version != expected {
		t.Fatalf("expected Goose version %d, got %d", expected, version)
	}
}

func appStoreConfig(t *testing.T) app_store.Config {
	t.Helper()
	return app_store.Config{
		Host:     requiredEnvironment(t, "APP_DB_HOST"),
		Port:     environmentPort(t, "APP_DB_PORT"),
		Database: requiredEnvironment(t, "APP_DB_NAME"),
		User:     requiredEnvironment(t, "APP_DB_USER"),
		Password: os.Getenv("APP_DB_PASSWORD"),
		SSLMode:  environmentOr("APP_DB_SSLMODE", "disable"),
	}
}

func keyStoreConfig(t *testing.T) key_store.Config {
	t.Helper()
	return key_store.Config{
		Host:     requiredEnvironment(t, "KEY_DB_HOST"),
		Port:     environmentPort(t, "KEY_DB_PORT"),
		Database: requiredEnvironment(t, "KEY_DB_NAME"),
		User:     requiredEnvironment(t, "KEY_DB_USER"),
		Password: os.Getenv("KEY_DB_PASSWORD"),
		SSLMode:  environmentOr("KEY_DB_SSLMODE", "disable"),
	}
}

func environmentPort(t *testing.T, name string) int {
	t.Helper()

	value := environmentOr(name, "5432")
	port, err := strconv.Atoi(value)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return port
}

func requiredEnvironment(t *testing.T, name string) string {
	t.Helper()

	value := os.Getenv(name)
	if value == "" {
		t.Fatalf("%s is required", name)
	}
	return value
}

func environmentOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
