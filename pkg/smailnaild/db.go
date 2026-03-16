package smailnaild

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	claysql "github.com/go-go-golems/clay/pkg/sql"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/jmoiron/sqlx"
)

const DefaultSQLiteDBPath = "smailnaild.sqlite"

const (
	bootstrapBaselineVersion = 1
	currentSchemaVersion     = 5
)

type DatabaseInfo struct {
	Driver string `json:"driver"`
	Target string `json:"target"`
	Mode   string `json:"mode"`
}

func OpenApplicationDB(ctx context.Context, parsedValues *values.Values) (*sqlx.DB, DatabaseInfo, error) {
	config, err := LoadDatabaseConfig(parsedValues)
	if err != nil {
		return nil, DatabaseInfo{}, err
	}

	db, err := config.Connect(ctx)
	if err != nil {
		return nil, DatabaseInfo{}, err
	}

	return db, DatabaseInfoFromConfig(config), nil
}

func LoadDatabaseConfig(parsedValues *values.Values) (*claysql.DatabaseConfig, error) {
	config, err := claysql.NewConfigFromRawParsedLayers(parsedValues)
	if err != nil {
		return nil, err
	}

	ApplyDatabaseDefaults(config)
	return config, nil
}

func ApplyDatabaseDefaults(config *claysql.DatabaseConfig) {
	if config == nil {
		return
	}

	if !hasExplicitDatabaseConfiguration(config) {
		config.Type = "sqlite"
		config.Driver = "sqlite3"
		config.Database = DefaultSQLiteDBPath
		config.Host = ""
		config.User = ""
		config.Password = ""
		config.Port = 0
		config.Schema = ""
		config.SSLDisable = true
		return
	}

	if config.DSN == "" && config.Database == "" && isSQLiteConfig(config) {
		config.Database = DefaultSQLiteDBPath
	}
}

func hasExplicitDatabaseConfiguration(config *claysql.DatabaseConfig) bool {
	if config == nil {
		return false
	}

	return config.UseDbtProfiles ||
		config.DbtProfile != "" ||
		config.DSN != "" ||
		config.Database != "" ||
		config.Host != "" ||
		config.User != "" ||
		config.Password != "" ||
		config.Driver != "" ||
		(config.Type != "" && config.Type != "mysql")
}

func isSQLiteConfig(config *claysql.DatabaseConfig) bool {
	switch strings.ToLower(config.Type) {
	case "sqlite", "sqlite3":
		return true
	}

	switch strings.ToLower(config.Driver) {
	case "sqlite", "sqlite3":
		return true
	}

	return false
}

func DatabaseInfoFromConfig(config *claysql.DatabaseConfig) DatabaseInfo {
	if config == nil {
		return DatabaseInfo{}
	}

	driver := normalizedDriver(config)
	mode := "structured"
	target := config.Database

	switch {
	case config.UseDbtProfiles:
		mode = "dbt"
		target = config.DbtProfile
	case config.DSN != "":
		mode = "dsn"
		target = redactDSN(config.DSN)
	case target == "":
		target = DefaultSQLiteDBPath
	}

	return DatabaseInfo{
		Driver: driver,
		Target: target,
		Mode:   mode,
	}
}

func BootstrapApplicationDB(ctx context.Context, db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS app_metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return err
	}

	version, err := schemaVersion(ctx, db)
	if err != nil {
		return err
	}
	if version == 0 {
		if err := setSchemaVersion(ctx, db, bootstrapBaselineVersion); err != nil {
			return err
		}
		version = bootstrapBaselineVersion
	}

	for _, migration := range schemaMigrations() {
		if migration.version <= version {
			continue
		}
		for _, statement := range migration.statements {
			if _, err := db.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("apply schema version %d: %w", migration.version, err)
			}
		}
		if err := setSchemaVersion(ctx, db, migration.version); err != nil {
			return err
		}
		version = migration.version
	}

	return nil
}

func PingDatabase(ctx context.Context, db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return db.PingContext(pingCtx)
}

type schemaMigration struct {
	version    int
	statements []string
}

func schemaMigrations() []schemaMigration {
	return []schemaMigration{
		{
			version: 2,
			statements: []string{
				`CREATE TABLE IF NOT EXISTS imap_accounts (
					id TEXT PRIMARY KEY,
					user_id TEXT NOT NULL,
					label TEXT NOT NULL,
					provider_hint TEXT NOT NULL DEFAULT '',
					server TEXT NOT NULL,
					port INTEGER NOT NULL,
					username TEXT NOT NULL,
					mailbox_default TEXT NOT NULL,
					insecure BOOLEAN NOT NULL DEFAULT FALSE,
					auth_kind TEXT NOT NULL,
					secret_ciphertext TEXT NOT NULL,
					secret_nonce TEXT NOT NULL,
					secret_key_id TEXT NOT NULL,
					is_default BOOLEAN NOT NULL DEFAULT FALSE,
					mcp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`,
			},
		},
		{
			version: 3,
			statements: []string{
				`CREATE TABLE IF NOT EXISTS imap_account_tests (
					id TEXT PRIMARY KEY,
					imap_account_id TEXT NOT NULL,
					test_mode TEXT NOT NULL,
					success BOOLEAN NOT NULL,
					tcp_ok BOOLEAN NOT NULL,
					login_ok BOOLEAN NOT NULL,
					mailbox_select_ok BOOLEAN NOT NULL,
					list_ok BOOLEAN NOT NULL,
					sample_fetch_ok BOOLEAN NOT NULL,
					write_probe_ok BOOLEAN,
					warning_code TEXT NOT NULL DEFAULT '',
					error_code TEXT NOT NULL DEFAULT '',
					error_message TEXT NOT NULL DEFAULT '',
					details_json TEXT NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`,
			},
		},
		{
			version: 4,
			statements: []string{
				`CREATE TABLE IF NOT EXISTS rules (
					id TEXT PRIMARY KEY,
					user_id TEXT NOT NULL,
					imap_account_id TEXT NOT NULL,
					name TEXT NOT NULL,
					description TEXT NOT NULL DEFAULT '',
					status TEXT NOT NULL,
					source_kind TEXT NOT NULL,
					rule_yaml TEXT NOT NULL,
					last_preview_count INTEGER NOT NULL DEFAULT 0,
					last_run_at TIMESTAMP,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`,
			},
		},
		{
			version: 5,
			statements: []string{
				`CREATE TABLE IF NOT EXISTS rule_runs (
					id TEXT PRIMARY KEY,
					rule_id TEXT NOT NULL,
					user_id TEXT NOT NULL,
					imap_account_id TEXT NOT NULL,
					mode TEXT NOT NULL,
					matched_count INTEGER NOT NULL,
					action_summary_json TEXT NOT NULL,
					sample_results_json TEXT NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				)`,
			},
		},
	}
}

func schemaVersion(ctx context.Context, db *sqlx.DB) (int, error) {
	var raw string
	err := db.GetContext(ctx, &raw, `SELECT value FROM app_metadata WHERE key = 'schema_version'`)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	switch raw {
	case "":
		return 0, nil
	case "1":
		return 1, nil
	case "2":
		return 2, nil
	case "3":
		return 3, nil
	case "4":
		return 4, nil
	case "5":
		return 5, nil
	default:
		return 0, fmt.Errorf("unsupported schema version %q", raw)
	}
}

func setSchemaVersion(ctx context.Context, db *sqlx.DB, version int) error {
	query := db.Rebind(`INSERT INTO app_metadata (key, value, updated_at)
	VALUES ('schema_version', ?, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = CURRENT_TIMESTAMP`)
	if _, err := db.ExecContext(ctx, query, fmt.Sprintf("%d", version)); err != nil {
		return err
	}
	return nil
}

func normalizedDriver(config *claysql.DatabaseConfig) string {
	if config == nil {
		return ""
	}

	driver := strings.ToLower(strings.TrimSpace(config.Driver))
	if driver != "" {
		switch driver {
		case "postgres", "postgresql", "pg":
			return "pgx"
		case "sqlite":
			return "sqlite3"
		default:
			return driver
		}
	}

	switch strings.ToLower(strings.TrimSpace(config.Type)) {
	case "postgres", "postgresql", "pg":
		return "pgx"
	case "sqlite":
		return "sqlite3"
	default:
		return strings.TrimSpace(config.Type)
	}
}

func redactDSN(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "dsn"
	}

	if parsed.User != nil {
		username := parsed.User.Username()
		if username != "" {
			parsed.User = url.UserPassword(username, "***")
		} else {
			parsed.User = url.User("***")
		}
	}

	return parsed.String()
}
