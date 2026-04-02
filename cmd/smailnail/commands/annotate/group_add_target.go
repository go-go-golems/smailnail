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

type GroupAddTargetCommand struct {
	*cmds.CommandDescription
}

type groupAddTargetSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	GroupID    string `glazed:"group-id"`
	TargetType string `glazed:"target-type"`
	TargetID   string `glazed:"target-id"`
}

func NewGroupAddTargetCommand() (*GroupAddTargetCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-group-add-target",
		"Group Add Target Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("group-id", fields.TypeString, fields.WithHelp("Group id to add a target to")),
			fields.New("target-type", fields.TypeString, fields.WithHelp("Target type")),
			fields.New("target-id", fields.TypeString, fields.WithHelp("Target id")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create group add-target section: %w", err)
	}
	return &GroupAddTargetCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add-target",
			cmds.WithShort("Add a target to a group"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *GroupAddTargetCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &groupAddTargetSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-group-add-target", settings); err != nil {
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

	if err := repo.AddGroupMember(ctx, annotatepkg.AddGroupMemberInput{
		GroupID:    settings.GroupID,
		TargetType: settings.TargetType,
		TargetID:   settings.TargetID,
	}); err != nil {
		return err
	}
	row := types.NewRow()
	row.Set("group_id", settings.GroupID)
	row.Set("target_type", settings.TargetType)
	row.Set("target_id", settings.TargetID)
	row.Set("status", "linked")
	return gp.AddRow(ctx, row)
}
