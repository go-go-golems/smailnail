package smailnaild

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	claysql "github.com/go-go-golems/clay/pkg/sql"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/jmoiron/sqlx"
)

const DefaultSQLiteDBPath = "smailnaild.sqlite"

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

	schemaStatements := []string{
		`CREATE TABLE IF NOT EXISTS app_metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO app_metadata (key, value, updated_at)
		VALUES ('schema_version', '1', CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP`,
	}

	for _, statement := range schemaStatements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
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
