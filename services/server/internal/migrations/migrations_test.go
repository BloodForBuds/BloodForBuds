package migrations

import (
	"testing"

	"github.com/pressly/goose/v3"
)

const expectedGooseVersion int64 = 1

func TestGooseVersion(t *testing.T) {
	migrations, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}

	latest, err := migrations.Last()
	if err != nil {
		t.Fatalf("get latest migration: %v", err)
	}
	if latest.Version != expectedGooseVersion {
		t.Fatalf("expected Goose version %d, got %d", expectedGooseVersion, latest.Version)
	}
}
