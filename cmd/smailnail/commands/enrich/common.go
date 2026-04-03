package enrich

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	enrichpkg "github.com/go-go-golems/smailnail/pkg/enrich"
	"github.com/go-go-golems/smailnail/pkg/mirror"
)

type enrichSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	AccountKey string `glazed:"account-key"`
	Mailbox    string `glazed:"mailbox"`
	Rebuild    bool   `glazed:"rebuild"`
	DryRun     bool   `glazed:"dry-run"`
}

func toOptions(settings enrichSettings) enrichpkg.Options {
	return enrichpkg.Options{
		AccountKey: settings.AccountKey,
		Mailbox:    settings.Mailbox,
		Rebuild:    settings.Rebuild,
		DryRun:     settings.DryRun,
	}
}

func openEnrichDB(ctx context.Context, sqlitePath string) (*sqlx.DB, func(), error) {
	store, err := mirror.OpenStore(sqlitePath)
	if err != nil {
		return nil, nil, err
	}

	bootstrapRoot := filepath.Join(filepath.Dir(sqlitePath), ".smailnail-enrich")
	if _, err := store.Bootstrap(ctx, bootstrapRoot); err != nil {
		_ = store.Close()
		return nil, nil, err
	}
	if err := store.Close(); err != nil {
		return nil, nil, errors.Wrap(err, "close bootstrapped store handle")
	}

	db, err := sqlx.Open("sqlite3", sqlitePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "open sqlite db for enrich command")
	}
	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup, nil
}

func buildMessageExistsClause(accountKey, mailbox string) (string, []any) {
	clauses := []string{"1 = 1"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(accountKey) != "" {
		clauses = append(clauses, "m.account_key = ?")
		args = append(args, accountKey)
	}
	if strings.TrimSpace(mailbox) != "" {
		clauses = append(clauses, "m.mailbox_name = ?")
		args = append(args, mailbox)
	}
	return strings.Join(clauses, " AND "), args
}

func requireSQLitePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("sqlite-path is required")
	}
	return nil
}
