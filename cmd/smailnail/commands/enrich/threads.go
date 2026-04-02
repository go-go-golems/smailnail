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

type ThreadsCommand struct {
	*cmds.CommandDescription
}

type threadsSettings struct {
	enrichSettings
}

func NewThreadsCommand() (*ThreadsCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}

	enrichSection, err := schema.NewSection(
		"enrich-threads",
		"Thread Enrichment Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("account-key", fields.TypeString, fields.WithHelp("Limit enrichment to one account key")),
			fields.New("mailbox", fields.TypeString, fields.WithHelp("Limit enrichment to one mailbox")),
			fields.New("rebuild", fields.TypeBool, fields.WithHelp("Reprocess all threads within scope"), fields.WithDefault(false)),
			fields.New("dry-run", fields.TypeBool, fields.WithHelp("Compute thread enrichment without writing changes"), fields.WithDefault(false)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread enrich section: %w", err)
	}

	return &ThreadsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"threads",
			cmds.WithShort("Reconstruct message threads in the local mirror database"),
			cmds.WithSections(glazedSection, enrichSection),
		),
	}, nil
}

func (c *ThreadsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &threadsSettings{}
	if err := parsedValues.DecodeSectionInto("enrich-threads", settings); err != nil {
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

	report, err := (&enrichpkg.ThreadEnricher{}).Enrich(ctx, db, toOptions(settings.enrichSettings))
	if err != nil {
		return err
	}

	row := types.NewRow()
	row.Set("messages_processed", report.MessagesProcessed)
	row.Set("threads_created", report.ThreadsCreated)
	row.Set("threads_updated", report.ThreadsUpdated)
	row.Set("elapsed_ms", report.ElapsedMS)
	return gp.AddRow(ctx, row)
}
