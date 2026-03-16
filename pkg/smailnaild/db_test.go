package smailnaild

import (
	"context"
	"database/sql"
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
	if value != "5" {
		t.Fatalf("expected schema version 5, got %q", value)
	}
}

func TestBootstrapApplicationDBCreatesPhaseOneAndTwoTables(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	if err := BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	expected := []string{
		"app_metadata",
		"imap_accounts",
		"imap_account_tests",
		"rules",
		"rule_runs",
	}
	for _, table := range expected {
		var name string
		err := db.Get(&name, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table)
		if err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestBootstrapApplicationDBMigratesLegacyVersionOneDatabase(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	_, err := db.Exec(`CREATE TABLE app_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	_, err = db.Exec(`INSERT INTO app_metadata (key, value, updated_at)
	VALUES ('schema_version', '1', CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("insert legacy version: %v", err)
	}

	if err := BootstrapApplicationDB(context.Background(), db); err != nil {
		t.Fatalf("BootstrapApplicationDB() error = %v", err)
	}

	var value string
	if err := db.Get(&value, `SELECT value FROM app_metadata WHERE key = 'schema_version'`); err != nil {
		t.Fatalf("failed to read upgraded schema_version: %v", err)
	}
	if value != "5" {
		t.Fatalf("expected upgraded schema version 5, got %q", value)
	}
}

func TestSchemaVersionReturnsZeroWhenMissing(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	_, err := db.Exec(`CREATE TABLE app_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create metadata: %v", err)
	}

	version, err := schemaVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("schemaVersion() error = %v", err)
	}
	if version != 0 {
		t.Fatalf("expected missing schema version to return 0, got %d", version)
	}
}

func TestSchemaVersionRejectsUnknownVersion(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	_, err := db.Exec(`CREATE TABLE app_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	_, err = db.Exec(`INSERT INTO app_metadata (key, value, updated_at)
	VALUES ('schema_version', '999', CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatalf("insert unknown version: %v", err)
	}

	_, err = schemaVersion(context.Background(), db)
	if err == nil {
		t.Fatal("expected schemaVersion() to fail for unknown version")
	}
}

func TestSchemaVersionHandlesNoRows(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	_, err := db.Exec(`CREATE TABLE app_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	if err := db.Get(new(string), `SELECT value FROM app_metadata WHERE key = 'schema_version'`); err == nil {
		t.Fatal("expected direct select to have no rows")
	} else if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}
