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

type AnnotationReviewCommand struct {
	*cmds.CommandDescription
}

type annotationReviewSettings struct {
	SQLitePath  string `glazed:"sqlite-path"`
	ID          string `glazed:"id"`
	ReviewState string `glazed:"review-state"`
}

func NewAnnotationReviewCommand() (*AnnotationReviewCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-annotation-review",
		"Annotation Review Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("id", fields.TypeString, fields.WithHelp("Annotation id to update")),
			fields.New("review-state", fields.TypeString, fields.WithHelp("New review state"), fields.WithDefault(annotatepkg.ReviewStateReviewed)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create annotation review section: %w", err)
	}
	return &AnnotationReviewCommand{
		CommandDescription: cmds.NewCommandDescription(
			"review",
			cmds.WithShort("Update the review state of an annotation"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *AnnotationReviewCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &annotationReviewSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-annotation-review", settings); err != nil {
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

	annotation, err := repo.UpdateAnnotationReviewState(ctx, settings.ID, settings.ReviewState)
	if err != nil {
		return err
	}
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
	return gp.AddRow(ctx, row)
}
