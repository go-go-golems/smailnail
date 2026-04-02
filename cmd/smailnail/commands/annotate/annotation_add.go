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

type AnnotationAddCommand struct {
	*cmds.CommandDescription
}

type annotationAddSettings struct {
	SQLitePath   string `glazed:"sqlite-path"`
	TargetType   string `glazed:"target-type"`
	TargetID     string `glazed:"target-id"`
	Tag          string `glazed:"tag"`
	NoteMarkdown string `glazed:"note"`
	SourceKind   string `glazed:"source-kind"`
	SourceLabel  string `glazed:"source-label"`
	AgentRunID   string `glazed:"agent-run-id"`
	ReviewState  string `glazed:"review-state"`
	CreatedBy    string `glazed:"created-by"`
}

func NewAnnotationAddCommand() (*AnnotationAddCommand, error) {
	glazedSection, err := settings.NewGlazedSection()
	if err != nil {
		return nil, fmt.Errorf("failed to create glazed section: %w", err)
	}
	section, err := schema.NewSection(
		"annotate-annotation-add",
		"Annotation Add Settings",
		schema.WithFields(
			fields.New("sqlite-path", fields.TypeString, fields.WithHelp("SQLite path for the local mirror store"), fields.WithDefault("smailnail-mirror.sqlite")),
			fields.New("target-type", fields.TypeString, fields.WithHelp("Target type such as message, thread, sender, domain, mailbox, or account")),
			fields.New("target-id", fields.TypeString, fields.WithHelp("Target identifier within the target type")),
			fields.New("tag", fields.TypeString, fields.WithHelp("Short free-form tag to attach")),
			fields.New("note", fields.TypeString, fields.WithHelp("Free-form markdown note to attach")),
			fields.New("source-kind", fields.TypeString, fields.WithHelp("Source kind: human, agent, heuristic, import"), fields.WithDefault(annotatepkg.SourceKindHuman)),
			fields.New("source-label", fields.TypeString, fields.WithHelp("Optional source label, such as the agent or workflow name")),
			fields.New("agent-run-id", fields.TypeString, fields.WithHelp("Optional agent run identifier that produced this annotation")),
			fields.New("review-state", fields.TypeString, fields.WithHelp("Review state: to_review, reviewed, dismissed")),
			fields.New("created-by", fields.TypeString, fields.WithHelp("Free-form creator label, such as manuel or triage-agent")),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create annotation add section: %w", err)
	}
	return &AnnotationAddCommand{
		CommandDescription: cmds.NewCommandDescription(
			"add",
			cmds.WithShort("Create an annotation on a target in the local mirror database"),
			cmds.WithSections(glazedSection, section),
		),
	}, nil
}

func (c *AnnotationAddCommand) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	settings := &annotationAddSettings{}
	if err := parsedValues.DecodeSectionInto("annotate-annotation-add", settings); err != nil {
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

	annotation, err := repo.CreateAnnotation(ctx, annotatepkg.CreateAnnotationInput{
		TargetType:   settings.TargetType,
		TargetID:     settings.TargetID,
		Tag:          settings.Tag,
		NoteMarkdown: settings.NoteMarkdown,
		SourceKind:   settings.SourceKind,
		SourceLabel:  settings.SourceLabel,
		AgentRunID:   settings.AgentRunID,
		ReviewState:  settings.ReviewState,
		CreatedBy:    settings.CreatedBy,
	})
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
