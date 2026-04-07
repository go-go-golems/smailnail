package mirror

import (
	"context"
	"testing"

	"github.com/go-go-golems/smailnail/pkg/annotate"
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
		"review_feedback",
		"review_feedback_targets",
		"review_guidelines",
		"run_guideline_links",
	}
	for _, table := range expected {
		var name string
		if err := db.Get(&name, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table); err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}
}

func TestBootstrapUpgradesLegacyVersion3DatabasesToReviewTables(t *testing.T) {
	db := sqlx.MustOpen("sqlite3", ":memory:")
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE mirror_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create mirror_metadata: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO mirror_metadata (key, value) VALUES ('schema_version', '3')`); err != nil {
		t.Fatalf("seed schema version: %v", err)
	}
	for _, statement := range annotate.SchemaMigrationV3CoreStatements() {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("seed v3 core statement %q: %v", statement, err)
		}
	}

	report, err := bootstrapSchema(ctx, db)
	if err != nil {
		t.Fatalf("bootstrapSchema() error = %v", err)
	}
	if report.SchemaVersion != currentSchemaVersion {
		t.Fatalf("expected schema version %d after upgrade, got %d", currentSchemaVersion, report.SchemaVersion)
	}

	for _, table := range []string{"review_feedback", "review_feedback_targets", "review_guidelines", "run_guideline_links"} {
		var name string
		if err := db.Get(&name, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table); err != nil {
			t.Fatalf("expected upgraded table %q to exist: %v", table, err)
		}
	}

	var version string
	if err := db.Get(&version, `SELECT value FROM mirror_metadata WHERE key = 'schema_version'`); err != nil {
		t.Fatalf("read upgraded schema version: %v", err)
	}
	if version != "4" {
		t.Fatalf("expected schema_version metadata to be 4, got %q", version)
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
