package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pressly/goose/v3"
)

const expectedGooseVersion int64 = 1

func TestMigrationsEmbedded(t *testing.T) {
	migrations, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		t.Fatalf("glob embedded migrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least one embedded migration")
	}
}

func TestGooseVersion(t *testing.T) {
	goose.SetBaseFS(migrationFiles)
	t.Cleanup(func() { goose.SetBaseFS(nil) })

	migrations, err := goose.CollectMigrations("migrations", 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect embedded migrations: %v", err)
	}

	latest, err := migrations.Last()
	if err != nil {
		t.Fatalf("get latest embedded migration: %v", err)
	}
	if latest.Version != expectedGooseVersion {
		t.Fatalf("expected Goose version %d, got %d", expectedGooseVersion, latest.Version)
	}
}

func TestHealthz(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	healthz(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}

	if body := response.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response body: %q", body)
	}
}
