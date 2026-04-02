package mirror

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-go-golems/smailnail/pkg/enrich"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	metadataTable        = "mirror_metadata"
	currentSchemaVersion = 2
)

type schemaMigration struct {
	version    int
	statements []string
}

func schemaMigrations() []schemaMigration {
	return []schemaMigration{
		{
			version: 1,
			statements: []string{
				`CREATE TABLE IF NOT EXISTS mailbox_sync_state (
					account_key TEXT NOT NULL,
					mailbox_name TEXT NOT NULL,
					uidvalidity INTEGER NOT NULL,
					highest_uid INTEGER NOT NULL DEFAULT 0,
					last_uidnext INTEGER NOT NULL DEFAULT 0,
					last_sync_at TIMESTAMP,
					status TEXT NOT NULL DEFAULT 'active',
					PRIMARY KEY (account_key, mailbox_name)
				)`,
				`CREATE TABLE IF NOT EXISTS messages (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					account_key TEXT NOT NULL,
					mailbox_name TEXT NOT NULL,
					uidvalidity INTEGER NOT NULL,
					uid INTEGER NOT NULL,
					message_id TEXT NOT NULL DEFAULT '',
					internal_date TEXT NOT NULL DEFAULT '',
					sent_date TEXT NOT NULL DEFAULT '',
					subject TEXT NOT NULL DEFAULT '',
					from_summary TEXT NOT NULL DEFAULT '',
					to_summary TEXT NOT NULL DEFAULT '',
					cc_summary TEXT NOT NULL DEFAULT '',
					size_bytes INTEGER NOT NULL DEFAULT 0,
					flags_json TEXT NOT NULL DEFAULT '[]',
					headers_json TEXT NOT NULL DEFAULT '{}',
					parts_json TEXT NOT NULL DEFAULT '[]',
					body_text TEXT NOT NULL DEFAULT '',
					body_html TEXT NOT NULL DEFAULT '',
					search_text TEXT NOT NULL DEFAULT '',
					raw_path TEXT NOT NULL,
					raw_sha256 TEXT NOT NULL DEFAULT '',
					has_attachments BOOLEAN NOT NULL DEFAULT FALSE,
					remote_deleted BOOLEAN NOT NULL DEFAULT FALSE,
					first_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					last_synced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					UNIQUE (account_key, mailbox_name, uidvalidity, uid)
				)`,
				`CREATE INDEX IF NOT EXISTS idx_messages_mailbox_uid
					ON messages(account_key, mailbox_name, uidvalidity, uid)`,
				`CREATE INDEX IF NOT EXISTS idx_messages_message_id
					ON messages(message_id)`,
				`CREATE INDEX IF NOT EXISTS idx_messages_dates
					ON messages(internal_date, sent_date)`,
			},
		},
		{
			version:    2,
			statements: enrich.SchemaMigrationV2Statements(),
		},
	}
}

func bootstrapSchema(ctx context.Context, db *sqlx.DB) (BootstrapReport, error) {
	if db == nil {
		return BootstrapReport{}, fmt.Errorf("database is nil")
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS mirror_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return BootstrapReport{}, errors.Wrap(err, "create mirror metadata table")
	}

	version, err := schemaVersion(ctx, db)
	if err != nil {
		return BootstrapReport{}, err
	}
	if version == 0 {
		if err := setSchemaVersion(ctx, db, 0); err != nil {
			return BootstrapReport{}, err
		}
	}

	for _, migration := range schemaMigrations() {
		if migration.version <= version {
			continue
		}
		for _, statement := range migration.statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				if isIgnorableMigrationError(err) {
					continue
				}
				return BootstrapReport{}, fmt.Errorf("apply mirror schema version %d: %w", migration.version, err)
			}
		}
		if err := setSchemaVersion(ctx, db, migration.version); err != nil {
			return BootstrapReport{}, err
		}
		version = migration.version
	}

	ftsAvailable, ftsStatus, err := bootstrapFTS(ctx, db)
	if err != nil {
		return BootstrapReport{}, err
	}

	return BootstrapReport{
		SearchMode:    SearchModeFTS5,
		FTSAvailable:  ftsAvailable,
		FTSStatus:     ftsStatus,
		SchemaVersion: version,
	}, nil
}

func bootstrapFTS(ctx context.Context, db *sqlx.DB) (bool, string, error) {
	_, err := db.ExecContext(ctx, `CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
		account_key,
		mailbox_name,
		subject,
		from_summary,
		to_summary,
		cc_summary,
		body_text,
		body_html,
		search_text
	)`)
	if err != nil {
		return false, "", fmt.Errorf("fts5 is required but unavailable: %w", err)
	}

	if err := setMetadataValue(ctx, db, "fts5_status", "available"); err != nil {
		return false, "", err
	}
	return true, "available", nil
}

func schemaVersion(ctx context.Context, db *sqlx.DB) (int, error) {
	var raw string
	err := db.GetContext(ctx, &raw, `SELECT value FROM mirror_metadata WHERE key = 'schema_version'`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, errors.Wrap(err, "read mirror schema version")
	}
	switch raw {
	case "", "0":
		return 0, nil
	case "1":
		return 1, nil
	case "2":
		return 2, nil
	default:
		return 0, fmt.Errorf("unsupported mirror schema version %q", raw)
	}
}

func setSchemaVersion(ctx context.Context, db *sqlx.DB, version int) error {
	return setMetadataValue(ctx, db, "schema_version", fmt.Sprintf("%d", version))
}

func setMetadataValue(ctx context.Context, db *sqlx.DB, key, value string) error {
	query := db.Rebind(`INSERT INTO mirror_metadata (key, value, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = CURRENT_TIMESTAMP`)
	if _, err := db.ExecContext(ctx, query, key, value); err != nil {
		return errors.Wrapf(err, "set mirror metadata %s", key)
	}
	return nil
}

func isIgnorableMigrationError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "duplicate column name")
}
