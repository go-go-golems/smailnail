package enrich

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"

	enrichpkg "github.com/go-go-golems/smailnail/pkg/enrich"
)

type SendersCommand struct {
	*cmds.CommandDescription
}

type sendersSettings struct {
	enrichSettings
	ShowPrivateRelay bool `glazed:"show-private-relay"`
}

func NewSendersCommand() (*SendersCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	enrichSection, err := schema.NewSection(
		"enrich-senders",
		"Sender Enrichment Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("account-key", fields.TypeString, fields.WithHelp("Limit enrichment to one account key")),
			fields.New("mailbox", fields.TypeString, fields.WithHelp("Limit enrichment to one mailbox")),
			fields.New("rebuild", fields.TypeBool, fields.WithHelp("Reprocess all sender rows within scope"), fields.WithDefault(false)),
			fields.New("dry-run", fields.TypeBool, fields.WithHelp("Compute sender enrichment without writing changes"), fields.WithDefault(false)),
			fields.New("show-private-relay", fields.TypeBool, fields.WithHelp("Emit one row per private relay sender instead of a summary row"), fields.WithDefault(false)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender enrich section: %w", err)
	}

	return &SendersCommand{
		CommandDescription: cmds.NewCommandDescription(
			"senders",
			cmds.WithShort("Normalize senders in the local mirror database"),
			cmds.WithSections(glazedSection, enrichSection),
		),
	}, nil
}

func (c *SendersCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &sendersSettings{}
	if err := parsedValues.DecodeSectionInto("enrich-senders", settings); err != nil {
		return err
	}
	if err := requireSQLitePath(settings.SQLitePath); err != nil {
		return err
	}

	db, cleanup, err := openEnrichDB(ctx, settings.SQLitePath)
	if err != nil {
		return err
	}
	defer cleanup()

	report, err := (&enrichpkg.SenderEnricher{}).Enrich(ctx, db, toOptions(settings.enrichSettings))
	if err != nil {
		return err
	}

	if settings.ShowPrivateRelay && !settings.DryRun {
		existsClause, args := buildMessageExistsClause(settings.AccountKey, settings.Mailbox)
		query := `SELECT
			email,
			display_name,
			domain,
			is_private_relay,
			relay_display_domain,
			msg_count,
			first_seen_date,
			last_seen_date
		FROM senders s
		WHERE is_private_relay = TRUE
			AND EXISTS (
				SELECT 1
				FROM messages m
				WHERE m.sender_email = s.email
					AND ` + existsClause + `
			)
		ORDER BY msg_count DESC, email`

		rows, err := db.QueryxContext(ctx, db.Rebind(query), args...)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			values := map[string]any{}
			if err := rows.MapScan(values); err != nil {
				return err
			}
			row := types.NewRow()
			for key, value := range values {
				row.Set(key, value)
			}
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
		return rows.Err()
	}

	row := types.NewRow()
	row.Set("senders_created", report.SendersCreated)
	row.Set("senders_updated", report.SendersUpdated)
	row.Set("messages_tagged", report.MessagesTagged)
	row.Set("private_relay_count", report.PrivateRelayCount)
	row.Set("elapsed_ms", report.ElapsedMS)
	return gp.AddRow(ctx, row)
}
