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

type AnnotationListCommand struct {
	*cmds.CommandDescription
}

type annotationListSettings struct {
	SQLitePath  string `glazed:"sqlite-path"`
	TargetType  string `glazed:"target-type"`
	TargetID    string `glazed:"target-id"`
	Tag         string `glazed:"tag"`
	ReviewState string `glazed:"review-state"`
	SourceKind  string `glazed:"source-kind"`
	Limit       int    `glazed:"limit"`
}

func NewAnnotationListCommand() (*AnnotationListCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-annotation-list",
		"Annotation List Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("target-type", fields.TypeString, fields.WithHelp("Filter by target type")),
			fields.New("target-id", fields.TypeString, fields.WithHelp("Filter by target id")),
			fields.New("tag", fields.TypeString, fields.WithHelp("Filter by tag")),
			fields.New("review-state", fields.TypeString, fields.WithHelp("Filter by review state")),
			fields.New("source-kind", fields.TypeString, fields.WithHelp("Filter by source kind")),
			fields.New("limit", fields.TypeInteger, fields.WithHelp("Maximum number of rows to emit (0 means no limit)"), fields.WithDefault(100)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create annotation list section: %w", err)
	}
	return &AnnotationListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List annotations from the local mirror database"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *AnnotationListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &annotationListSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-annotation-list", settings); err != nil {
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

	annotations, err := repo.ListAnnotations(ctx, annotatepkg.ListAnnotationsFilter{
		TargetType:  settings.TargetType,
		TargetID:    settings.TargetID,
		Tag:         settings.Tag,
		ReviewState: settings.ReviewState,
		SourceKind:  settings.SourceKind,
		Limit:       settings.Limit,
	})
	if err != nil {
		return err
	}
	for _, annotation := range annotations {
		row := types.NewRow()
		row.Set("id", annotation.ID)
		row.Set("target_type", annotation.TargetType)
		row.Set("target_id", annotation.TargetID)
		row.Set("tag", annotation.Tag)
		row.Set("note_markdown", annotation.NoteMarkdown)
		row.Set("source_kind", annotation.SourceKind)
		row.Set("source_label", annotation.SourceLabel)
		row.Set("agent_run_id", annotation.AgentRunID)
		row.Set("review_state", annotation.ReviewState)
		row.Set("created_by", annotation.CreatedBy)
		row.Set("created_at", annotation.CreatedAt)
		row.Set("updated_at", annotation.UpdatedAt)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
