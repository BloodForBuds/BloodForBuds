package key_store

import (
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestGooseVersion(t *testing.T) {
	db, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatalf("open test database handle: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	provider, err := newMigrationProvider(db)
	if err != nil {
		t.Fatalf("create migration provider: %v", err)
	}
	sources := provider.ListSources()
	if len(sources) == 0 {
		t.Fatal("expected at least one key store migration")
	}

	latest := sources[len(sources)-1]
	if latest.Version != currentMigrationVersion {
		t.Fatalf("expected key store Goose version %d, got %d", currentMigrationVersion, latest.Version)
	}
	if latest.Type != goose.TypeSQL {
		t.Fatalf("expected key store migration to be SQL, got %s", latest.Type)
	}
}
