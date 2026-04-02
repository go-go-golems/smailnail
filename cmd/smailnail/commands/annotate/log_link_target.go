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

type LogLinkTargetCommand struct {
	*cmds.CommandDescription
}

type logLinkTargetSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	LogID      string `glazed:"log-id"`
	TargetType string `glazed:"target-type"`
	TargetID   string `glazed:"target-id"`
}

func NewLogLinkTargetCommand() (*LogLinkTargetCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-log-link-target",
		"Log Link Target Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("log-id", fields.TypeString, fields.WithHelp("Log id to link")),
			fields.New("target-type", fields.TypeString, fields.WithHelp("Target type")),
			fields.New("target-id", fields.TypeString, fields.WithHelp("Target id")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log link-target section: %w", err)
	}
	return &LogLinkTargetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"link-target",
			cmds.WithShort("Link a log entry to a target"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *LogLinkTargetCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &logLinkTargetSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-log-link-target", settings); err != nil {
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

	if err := repo.LinkLogTarget(ctx, annotatepkg.LinkLogTargetInput{
		LogID:      settings.LogID,
		TargetType: settings.TargetType,
		TargetID:   settings.TargetID,
	}); err != nil {
		return err
	}
	row := types.NewRow()
	row.Set("log_id", settings.LogID)
	row.Set("target_type", settings.TargetType)
	row.Set("target_id", settings.TargetID)
	row.Set("status", "linked")
	return gp.AddRow(ctx, row)
}
