package annotate

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
	annotatepkg "github.com/go-go-golems/smailnail/pkg/annotate"
)

type LogListCommand struct {
	*cmds.CommandDescription
}

type logListSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	SourceKind string `glazed:"source-kind"`
	AgentRunID string `glazed:"agent-run-id"`
	Limit      int    `glazed:"limit"`
}

func NewLogListCommand() (*LogListCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-log-list",
		"Log List Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("source-kind", fields.TypeString, fields.WithHelp("Filter by source kind")),
			fields.New("agent-run-id", fields.TypeString, fields.WithHelp("Filter by agent run id")),
			fields.New("limit", fields.TypeInteger, fields.WithHelp("Maximum number of rows to emit (0 means no limit)"), fields.WithDefault(100)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log list section: %w", err)
	}
	return &LogListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List annotation log entries"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *LogListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &logListSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-log-list", settings); err != nil {
		return err
	}
	if err := requireSQLitePath(settings.SQLitePath); err != nil {
		return err
	}
	repo, cleanup, err := openAnnotateRepo(ctx, settings.SQLitePath)
	if err != nil {
		return err
	}
	defer cleanup()

	logs, err := repo.ListLogs(ctx, annotatepkg.ListLogsFilter{
		SourceKind: settings.SourceKind,
		AgentRunID: settings.AgentRunID,
		Limit:      settings.Limit,
	})
	if err != nil {
		return err
	}
	for _, logEntry := range logs {
		row := types.NewRow()
		row.Set("id", logEntry.ID)
		row.Set("log_kind", logEntry.LogKind)
		row.Set("title", logEntry.Title)
		row.Set("body_markdown", logEntry.BodyMarkdown)
		row.Set("source_kind", logEntry.SourceKind)
		row.Set("source_label", logEntry.SourceLabel)
		row.Set("agent_run_id", logEntry.AgentRunID)
		row.Set("created_by", logEntry.CreatedBy)
		row.Set("created_at", logEntry.CreatedAt)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
