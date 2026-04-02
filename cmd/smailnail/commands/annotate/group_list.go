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

type GroupListCommand struct {
	*cmds.CommandDescription
}

type groupListSettings struct {
	SQLitePath  string `glazed:"sqlite-path"`
	ReviewState string `glazed:"review-state"`
	SourceKind  string `glazed:"source-kind"`
	Limit       int    `glazed:"limit"`
}

func NewGroupListCommand() (*GroupListCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-group-list",
		"Group List Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("review-state", fields.TypeString, fields.WithHelp("Filter by review state")),
			fields.New("source-kind", fields.TypeString, fields.WithHelp("Filter by source kind")),
			fields.New("limit", fields.TypeInteger, fields.WithHelp("Maximum number of rows to emit (0 means no limit)"), fields.WithDefault(100)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create group list section: %w", err)
	}
	return &GroupListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List target groups"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *GroupListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &groupListSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-group-list", settings); err != nil {
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

	groups, err := repo.ListGroups(ctx, annotatepkg.ListGroupsFilter{
		ReviewState: settings.ReviewState,
		SourceKind:  settings.SourceKind,
		Limit:       settings.Limit,
	})
	if err != nil {
		return err
	}
	for _, group := range groups {
		row := types.NewRow()
		row.Set("id", group.ID)
		row.Set("name", group.Name)
		row.Set("description", group.Description)
		row.Set("source_kind", group.SourceKind)
		row.Set("source_label", group.SourceLabel)
		row.Set("agent_run_id", group.AgentRunID)
		row.Set("review_state", group.ReviewState)
		row.Set("created_by", group.CreatedBy)
		row.Set("created_at", group.CreatedAt)
		row.Set("updated_at", group.UpdatedAt)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
