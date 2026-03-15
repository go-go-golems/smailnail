package smailnaild

import (
	"context"
	"testing"

	claysql "github.com/go-go-golems/clay/pkg/sql"
	"github.com/jmoiron/sqlx"
)

func TestApplyDatabaseDefaultsDefaultsToSQLite(t *testing.T) {
	config := &claysql.DatabaseConfig{}

	ApplyDatabaseDefaults(config)

	if config.Type != "sqlite" {
		t.Fatalf("expected sqlite default type, got %q", config.Type)
	}
	if config.Driver != "sqlite3" {
		t.Fatalf("expected sqlite3 default driver, got %q", config.Driver)
	}
	if config.Database != DefaultSQLiteDBPath {
		t.Fatalf("expected default sqlite path %q, got %q", DefaultSQLiteDBPath, config.Database)
	}
}

func TestApplyDatabaseDefaultsPreservesExplicitPostgresConfig(t *testing.T) {
	config := &claysql.DatabaseConfig{
		Type:     "postgres",
		Host:     "db.internal",
		Database: "smailnail",
		User:     "app",
	}

	ApplyDatabaseDefaults(config)

	if config.Type != "postgres" {
		t.Fatalf("expected postgres config to be preserved, got %q", config.Type)
	}
	if config.Database != "smailnail" {
		t.Fatalf("expected database to stay set, got %q", config.Database)
	}
}

func TestBootstrapApplicationDBCreatesMetadata(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	var value string
	if err := db.Get(&value, `SELECT value FROM app_metadata WHERE key = 'schema_version'`); err != nil {
		t.Fatalf("failed to read schema_version: %v", err)
	}
	if value != "1" {
		t.Fatalf("expected schema version 1, got %q", value)
	}
}
