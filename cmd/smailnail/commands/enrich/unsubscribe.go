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

type UnsubscribeCommand struct {
	*cmds.CommandDescription
}

type unsubscribeSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	AccountKey string `glazed:"account-key"`
	Mailbox    string `glazed:"mailbox"`
	Rebuild    bool   `glazed:"rebuild"`
	DryRun     bool   `glazed:"dry-run"`
	EmitLinks  bool   `glazed:"emit-links"`
}

func NewUnsubscribeCommand() (*UnsubscribeCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	enrichSection, err := schema.NewSection(
		"enrich-unsubscribe",
		"Unsubscribe Enrichment Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("account-key", fields.TypeString, fields.WithHelp("Limit enrichment to one account key")),
			fields.New("mailbox", fields.TypeString, fields.WithHelp("Limit enrichment to one mailbox")),
			fields.New("rebuild", fields.TypeBool, fields.WithHelp("Reprocess unsubscribe rows within scope"), fields.WithDefault(false)),
			fields.New("dry-run", fields.TypeBool, fields.WithHelp("Compute unsubscribe enrichment without writing changes"), fields.WithDefault(false)),
			fields.New("emit-links", fields.TypeBool, fields.WithHelp("Emit one row per sender with unsubscribe links instead of a summary row"), fields.WithDefault(false)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create unsubscribe enrich section: %w", err)
	}

	return &UnsubscribeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"unsubscribe",
			cmds.WithShort("Extract List-Unsubscribe links into normalized sender rows"),
			cmds.WithSections(glazedSection, enrichSection),
		),
	}, nil
}

func (c *UnsubscribeCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &unsubscribeSettings{}
	if err := parsedValues.DecodeSectionInto("enrich-unsubscribe", settings); err != nil {
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

	report, err := (&enrichpkg.UnsubscribeEnricher{}).Enrich(ctx, db, toOptions(enrichSettings{
		SQLitePath: settings.SQLitePath,
		AccountKey: settings.AccountKey,
		Mailbox:    settings.Mailbox,
		Rebuild:    settings.Rebuild,
		DryRun:     settings.DryRun,
	}))
	if err != nil {
		return err
	}

	if settings.EmitLinks && !settings.DryRun {
		existsClause, args := buildMessageExistsClause(settings.AccountKey, settings.Mailbox)
		query := `SELECT
			email,
			display_name,
			domain,
			msg_count,
			unsubscribe_mailto,
			unsubscribe_http,
			has_list_unsubscribe,
			EXISTS (
				SELECT 1
				FROM messages m2
				WHERE m2.sender_email = s.email
					AND m2.headers_json LIKE '%List-Unsubscribe-Post%'
					AND m2.headers_json LIKE '%One-Click%'
			) AS one_click
		FROM senders s
		WHERE has_list_unsubscribe = TRUE
			AND EXISTS (
				SELECT 1
				FROM messages m
				WHERE m.sender_email = s.email
					AND ` + existsClause + `
			)
		ORDER BY email`
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
	row.Set("senders_with_unsubscribe", report.SendersWithUnsubscribe)
	row.Set("mailto_links", report.MailtoLinks)
	row.Set("http_links", report.HTTPLinks)
	row.Set("one_click_links", report.OneClickLinks)
	row.Set("elapsed_ms", report.ElapsedMS)
	return gp.AddRow(ctx, row)
}
