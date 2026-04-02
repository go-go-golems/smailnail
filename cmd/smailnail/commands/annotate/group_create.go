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

type GroupCreateCommand struct {
	*cmds.CommandDescription
}

type groupCreateSettings struct {
	SQLitePath  string `glazed:"sqlite-path"`
	Name        string `glazed:"name"`
	Description string `glazed:"description"`
	SourceKind  string `glazed:"source-kind"`
	SourceLabel string `glazed:"source-label"`
	AgentRunID  string `glazed:"agent-run-id"`
	ReviewState string `glazed:"review-state"`
	CreatedBy   string `glazed:"created-by"`
}

func NewGroupCreateCommand() (*GroupCreateCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-group-create",
		"Group Create Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("name", fields.TypeString, fields.WithHelp("Group name")),
			fields.New("description", fields.TypeString, fields.WithHelp("Optional group description")),
			fields.New("source-kind", fields.TypeString, fields.WithHelp("Source kind: human, agent, heuristic, import"), fields.WithDefault(annotatepkg.SourceKindHuman)),
			fields.New("source-label", fields.TypeString, fields.WithHelp("Optional source label")),
			fields.New("agent-run-id", fields.TypeString, fields.WithHelp("Optional agent run id")),
			fields.New("review-state", fields.TypeString, fields.WithHelp("Review state for the group")),
			fields.New("created-by", fields.TypeString, fields.WithHelp("Free-form creator label")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create group create section: %w", err)
	}
	return &GroupCreateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a target group"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *GroupCreateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &groupCreateSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-group-create", settings); err != nil {
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

	group, err := repo.CreateGroup(ctx, annotatepkg.CreateGroupInput{
		Name:        settings.Name,
		Description: settings.Description,
		SourceKind:  settings.SourceKind,
		SourceLabel: settings.SourceLabel,
		AgentRunID:  settings.AgentRunID,
		ReviewState: settings.ReviewState,
		CreatedBy:   settings.CreatedBy,
	})
	if err != nil {
		return err
	}

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
	return gp.AddRow(ctx, row)
}
