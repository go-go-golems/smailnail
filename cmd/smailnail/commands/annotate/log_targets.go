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
)

type LogTargetsCommand struct {
	*cmds.CommandDescription
}

type logTargetsSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	LogID      string `glazed:"log-id"`
}

func NewLogTargetsCommand() (*LogTargetsCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-log-targets",
		"Log Targets Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("log-id", fields.TypeString, fields.WithHelp("Log id to inspect")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log targets section: %w", err)
	}
	return &LogTargetsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"targets",
			cmds.WithShort("List targets linked to a log entry"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *LogTargetsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &logTargetsSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-log-targets", settings); err != nil {
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

	targets, err := repo.ListLogTargets(ctx, settings.LogID)
	if err != nil {
		return err
	}
	for _, target := range targets {
		row := types.NewRow()
		row.Set("log_id", target.LogID)
		row.Set("target_type", target.TargetType)
		row.Set("target_id", target.TargetID)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
