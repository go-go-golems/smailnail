package mirror

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestBootstrapCreatesCoreTables(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	report, err := bootstrapSchema(context.Background(), db)
	if err != nil {
		t.Fatalf("bootstrapSchema() error = %v", err)
	}
	if report.SchemaVersion != currentSchemaVersion {
		t.Fatalf("expected schema version %d, got %d", currentSchemaVersion, report.SchemaVersion)
	}
	if !report.FTSAvailable {
		t.Fatal("expected FTS to be available")
	}
	if report.SearchMode != SearchModeFTS5 {
		t.Fatalf("expected search mode %q, got %q", SearchModeFTS5, report.SearchMode)
	}

	expected := []string{
		"mirror_metadata",
		"mailbox_sync_state",
		"messages",
		"annotations",
		"target_groups",
		"annotation_logs",
	}
	for _, table := range expected {
		var name string
		if err := db.Get(&name, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table); err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestBootstrapRecordsFTSAvailability(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	report, err := bootstrapSchema(context.Background(), db)
	if err != nil {
		t.Fatalf("bootstrapSchema() error = %v", err)
	}
	if report.FTSStatus != "available" {
		t.Fatalf("expected FTS status %q, got %q", "available", report.FTSStatus)
	}

	var status string
	if err := db.Get(&status, `SELECT value FROM mirror_metadata WHERE key = 'fts5_status'`); err != nil {
		t.Fatalf("expected fts5_status metadata: %v", err)
	}
	if status == "" {
		t.Fatal("expected non-empty fts5_status metadata")
	}
}
