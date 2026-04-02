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

type LogAddCommand struct {
	*cmds.CommandDescription
}

type logAddSettings struct {
	SQLitePath   string `glazed:"sqlite-path"`
	LogKind      string `glazed:"log-kind"`
	Title        string `glazed:"title"`
	BodyMarkdown string `glazed:"body"`
	SourceKind   string `glazed:"source-kind"`
	SourceLabel  string `glazed:"source-label"`
	AgentRunID   string `glazed:"agent-run-id"`
	CreatedBy    string `glazed:"created-by"`
}

func NewLogAddCommand() (*LogAddCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-log-add",
		"Log Add Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("log-kind", fields.TypeString, fields.WithHelp("Log kind"), fields.WithDefault("note")),
			fields.New("title", fields.TypeString, fields.WithHelp("Log title")),
			fields.New("body", fields.TypeString, fields.WithHelp("Log body markdown")),
			fields.New("source-kind", fields.TypeString, fields.WithHelp("Source kind"), fields.WithDefault(annotatepkg.SourceKindHuman)),
			fields.New("source-label", fields.TypeString, fields.WithHelp("Optional source label")),
			fields.New("agent-run-id", fields.TypeString, fields.WithHelp("Optional agent run id")),
			fields.New("created-by", fields.TypeString, fields.WithHelp("Free-form creator label")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log add section: %w", err)
	}
	return &LogAddCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add",
			cmds.WithShort("Create an annotation log entry"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *LogAddCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &logAddSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-log-add", settings); err != nil {
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

	logEntry, err := repo.CreateLog(ctx, annotatepkg.CreateLogInput{
		LogKind:      settings.LogKind,
		Title:        settings.Title,
		BodyMarkdown: settings.BodyMarkdown,
		SourceKind:   settings.SourceKind,
		SourceLabel:  settings.SourceLabel,
		AgentRunID:   settings.AgentRunID,
		CreatedBy:    settings.CreatedBy,
	})
	if err != nil {
		return err
	}

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
	return gp.AddRow(ctx, row)
}
