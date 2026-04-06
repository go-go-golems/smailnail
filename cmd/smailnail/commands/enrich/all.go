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

type AllCommand struct {
	*cmds.CommandDescription
}

type allSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	AccountKey string `glazed:"account-key"`
	Mailbox    string `glazed:"mailbox"`
	Rebuild    bool   `glazed:"rebuild"`
	DryRun     bool   `glazed:"dry-run"`
}

func NewAllCommand() (*AllCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	enrichSection, err := schema.NewSection(
		"enrich-all",
		"All Enrichment Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("account-key", fields.TypeString, fields.WithHelp("Limit enrichment to one account key")),
			fields.New("mailbox", fields.TypeString, fields.WithHelp("Limit enrichment to one mailbox")),
			fields.New("rebuild", fields.TypeBool, fields.WithHelp("Reprocess all enrichment rows within scope"), fields.WithDefault(false)),
			fields.New("dry-run", fields.TypeBool, fields.WithHelp("Compute enrichment without writing changes"), fields.WithDefault(false)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create enrich-all section: %w", err)
	}

	return &AllCommand{
		CommandDescription: cmds.NewCommandDescription(
			"all",
			cmds.WithShort("Run sender, thread, and unsubscribe enrichment in sequence"),
			cmds.WithSections(glazedSection, enrichSection),
		),
	}, nil
}

func (c *AllCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &allSettings{}
	if err := parsedValues.DecodeSectionInto("enrich-all", settings); err != nil {
		return err
	}
	if err := requireSQLitePath(settings.SQLitePath); err != nil {
		return err
	}

	report, err := enrichpkg.RunAll(ctx, settings.SQLitePath, toOptions(enrichSettings{
		SQLitePath: settings.SQLitePath,
		AccountKey: settings.AccountKey,
		Mailbox:    settings.Mailbox,
		Rebuild:    settings.Rebuild,
		DryRun:     settings.DryRun,
	}))
	if err != nil {
		return err
	}

	row := types.NewRow()
	row.Set("senders_created", report.Senders.SendersCreated)
	row.Set("senders_updated", report.Senders.SendersUpdated)
	row.Set("messages_tagged", report.Senders.MessagesTagged)
	row.Set("private_relay_count", report.Senders.PrivateRelayCount)
	row.Set("messages_processed", report.Threads.MessagesProcessed)
	row.Set("threads_created", report.Threads.ThreadsCreated)
	row.Set("threads_updated", report.Threads.ThreadsUpdated)
	row.Set("senders_with_unsubscribe", report.Unsubscribe.SendersWithUnsubscribe)
	row.Set("mailto_links", report.Unsubscribe.MailtoLinks)
	row.Set("http_links", report.Unsubscribe.HTTPLinks)
	row.Set("one_click_links", report.Unsubscribe.OneClickLinks)
	row.Set("elapsed_ms", report.ElapsedMS)
	return gp.AddRow(ctx, row)
}
