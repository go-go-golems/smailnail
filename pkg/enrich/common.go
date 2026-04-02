package enrich

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	metadataKeyEnrichSendersAt = "enrich_senders_at"
)

func normalizeOptions(opts Options) Options {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 1000
	}
	return opts
}

func buildMessageScopeClause(opts Options, alias string) (string, []any) {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	clauses := []string{"1 = 1"}
	args := make([]any, 0, 2)
	if strings.TrimSpace(opts.AccountKey) != "" {
		clauses = append(clauses, prefix+"account_key = ?")
		args = append(args, opts.AccountKey)
	}
	if strings.TrimSpace(opts.Mailbox) != "" {
		clauses = append(clauses, prefix+"mailbox_name = ?")
		args = append(args, opts.Mailbox)
	}

	return strings.Join(clauses, " AND "), args
}

func setMetadataValue(ctx context.Context, tx *sqlx.Tx, key, value string) error {
	if tx == nil {
		return errors.New("metadata transaction is nil")
	}

	query := tx.Rebind(`INSERT INTO mirror_metadata (key, value, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = CURRENT_TIMESTAMP`)
	if _, err := tx.ExecContext(ctx, query, key, value); err != nil {
		return errors.Wrapf(err, "set enrichment metadata %s", key)
	}

	return nil
}
