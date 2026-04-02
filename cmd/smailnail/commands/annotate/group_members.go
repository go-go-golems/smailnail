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

type GroupMembersCommand struct {
	*cmds.CommandDescription
}

type groupMembersSettings struct {
	SQLitePath string `glazed:"sqlite-path"`
	GroupID    string `glazed:"group-id"`
}

func NewGroupMembersCommand() (*GroupMembersCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-group-members",
		"Group Members Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("group-id", fields.TypeString, fields.WithHelp("Group id to inspect")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create group members section: %w", err)
	}
	return &GroupMembersCommand{
		CommandDescription: cmds.NewCommandDescription(
			"members",
			cmds.WithShort("List members of a group"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *GroupMembersCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &groupMembersSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-group-members", settings); err != nil {
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

	members, err := repo.ListGroupMembers(ctx, settings.GroupID)
	if err != nil {
		return err
	}
	for _, member := range members {
		row := types.NewRow()
		row.Set("group_id", member.GroupID)
		row.Set("target_type", member.TargetType)
		row.Set("target_id", member.TargetID)
		row.Set("added_at", member.AddedAt)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
