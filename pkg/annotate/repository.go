package annotate

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Repository struct {
	db    *sqlx.DB
	newID func() string
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db:    db,
		newID: uuid.NewString,
	}
}

func (r *Repository) CreateAnnotation(ctx context.Context, input CreateAnnotationInput) (*Annotation, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("annotation repository database is nil")
	}
	if strings.TrimSpace(input.TargetType) == "" {
		return nil, fmt.Errorf("target_type is required")
	}
	if strings.TrimSpace(input.TargetID) == "" {
		return nil, fmt.Errorf("target_id is required")
	}
	if strings.TrimSpace(input.Tag) == "" && strings.TrimSpace(input.NoteMarkdown) == "" {
		return nil, fmt.Errorf("tag or note_markdown is required")
	}

	id := r.newID()
	sourceKind := defaultSourceKind(input.SourceKind)
	reviewState := defaultReviewState(input.ReviewState, sourceKind)
	query := r.db.Rebind(`INSERT INTO annotations (
		id, target_type, target_id, tag, note_markdown, source_kind, source_label,
		agent_run_id, review_state, created_by, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if _, err := r.db.ExecContext(
		ctx,
		query,
		id,
		strings.TrimSpace(input.TargetType),
		strings.TrimSpace(input.TargetID),
		strings.TrimSpace(input.Tag),
		strings.TrimSpace(input.NoteMarkdown),
		sourceKind,
		strings.TrimSpace(input.SourceLabel),
		strings.TrimSpace(input.AgentRunID),
		reviewState,
		strings.TrimSpace(input.CreatedBy),
	); err != nil {
		return nil, errors.Wrap(err, "insert annotation")
	}

	return r.GetAnnotation(ctx, id)
}

func (r *Repository) GetAnnotation(ctx context.Context, id string) (*Annotation, error) {
	var ret Annotation
	query := r.db.Rebind(`SELECT * FROM annotations WHERE id = ?`)
	if err := r.db.GetContext(ctx, &ret, query, id); err != nil {
		return nil, errors.Wrap(err, "get annotation")
	}
	return &ret, nil
}

func (r *Repository) ListAnnotations(ctx context.Context, filter ListAnnotationsFilter) ([]Annotation, error) {
	query := `SELECT * FROM annotations WHERE 1 = 1`
	args := make([]any, 0, 6)
	if strings.TrimSpace(filter.TargetType) != "" {
		query += ` AND target_type = ?`
		args = append(args, strings.TrimSpace(filter.TargetType))
	}
	if strings.TrimSpace(filter.TargetID) != "" {
		query += ` AND target_id = ?`
		args = append(args, strings.TrimSpace(filter.TargetID))
	}
	if strings.TrimSpace(filter.Tag) != "" {
		query += ` AND tag = ?`
		args = append(args, strings.TrimSpace(filter.Tag))
	}
	if strings.TrimSpace(filter.ReviewState) != "" {
		query += ` AND review_state = ?`
		args = append(args, strings.TrimSpace(filter.ReviewState))
	}
	if strings.TrimSpace(filter.SourceKind) != "" {
		query += ` AND source_kind = ?`
		args = append(args, strings.TrimSpace(filter.SourceKind))
	}
	if strings.TrimSpace(filter.AgentRunID) != "" {
		query += ` AND agent_run_id = ?`
		args = append(args, strings.TrimSpace(filter.AgentRunID))
	}
	query += ` ORDER BY created_at DESC, id DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	ret := []Annotation{}
	if err := r.db.SelectContext(ctx, &ret, r.db.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "list annotations")
	}
	return ret, nil
}

func (r *Repository) UpdateAnnotationReviewState(ctx context.Context, id, reviewState string) (*Annotation, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	reviewState = defaultReviewState(reviewState, "")
	query := r.db.Rebind(`UPDATE annotations
		SET review_state = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`)
	result, err := r.db.ExecContext(ctx, query, reviewState, id)
	if err != nil {
		return nil, errors.Wrap(err, "update annotation review state")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "read annotation review update result")
	}
	if rows == 0 {
		return nil, fmt.Errorf("annotation %s not found", id)
	}
	return r.GetAnnotation(ctx, id)
}

func (r *Repository) BatchUpdateReviewState(ctx context.Context, ids []string, reviewState string) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids are required")
	}

	trimmedIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		trimmedIDs = append(trimmedIDs, id)
	}
	if len(trimmedIDs) == 0 {
		return fmt.Errorf("ids are required")
	}

	query, args, err := sqlx.In(`UPDATE annotations
		SET review_state = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id IN (?)`, defaultReviewState(reviewState, ""), trimmedIDs)
	if err != nil {
		return errors.Wrap(err, "build batch review update query")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin batch review update transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, tx.Rebind(query), args...); err != nil {
		return errors.Wrap(err, "batch update annotation review state")
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit batch review update transaction")
	}

	return nil
}

func (r *Repository) CreateGroup(ctx context.Context, input CreateGroupInput) (*TargetGroup, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("annotation repository database is nil")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	id := r.newID()
	sourceKind := defaultSourceKind(input.SourceKind)
	reviewState := defaultReviewState(input.ReviewState, sourceKind)
	query := r.db.Rebind(`INSERT INTO target_groups (
		id, name, description, source_kind, source_label, agent_run_id,
		review_state, created_by, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if _, err := r.db.ExecContext(
		ctx,
		query,
		id,
		strings.TrimSpace(input.Name),
		strings.TrimSpace(input.Description),
		sourceKind,
		strings.TrimSpace(input.SourceLabel),
		strings.TrimSpace(input.AgentRunID),
		reviewState,
		strings.TrimSpace(input.CreatedBy),
	); err != nil {
		return nil, errors.Wrap(err, "insert target group")
	}
	return r.GetGroup(ctx, id)
}

func (r *Repository) GetGroup(ctx context.Context, id string) (*TargetGroup, error) {
	var ret TargetGroup
	query := r.db.Rebind(`SELECT * FROM target_groups WHERE id = ?`)
	if err := r.db.GetContext(ctx, &ret, query, id); err != nil {
		return nil, errors.Wrap(err, "get target group")
	}
	return &ret, nil
}

func (r *Repository) ListGroups(ctx context.Context, filter ListGroupsFilter) ([]TargetGroup, error) {
	query := `SELECT * FROM target_groups WHERE 1 = 1`
	args := make([]any, 0, 3)
	if strings.TrimSpace(filter.ReviewState) != "" {
		query += ` AND review_state = ?`
		args = append(args, strings.TrimSpace(filter.ReviewState))
	}
	if strings.TrimSpace(filter.SourceKind) != "" {
		query += ` AND source_kind = ?`
		args = append(args, strings.TrimSpace(filter.SourceKind))
	}
	query += ` ORDER BY created_at DESC, id DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}
	ret := []TargetGroup{}
	if err := r.db.SelectContext(ctx, &ret, r.db.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "list target groups")
	}
	return ret, nil
}

func (r *Repository) AddGroupMember(ctx context.Context, input AddGroupMemberInput) error {
	if strings.TrimSpace(input.GroupID) == "" {
		return fmt.Errorf("group_id is required")
	}
	if strings.TrimSpace(input.TargetType) == "" {
		return fmt.Errorf("target_type is required")
	}
	if strings.TrimSpace(input.TargetID) == "" {
		return fmt.Errorf("target_id is required")
	}
	query := r.db.Rebind(`INSERT INTO target_group_members (
		group_id, target_type, target_id, added_at
	) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(group_id, target_type, target_id) DO NOTHING`)
	if _, err := r.db.ExecContext(
		ctx,
		query,
		strings.TrimSpace(input.GroupID),
		strings.TrimSpace(input.TargetType),
		strings.TrimSpace(input.TargetID),
	); err != nil {
		return errors.Wrap(err, "insert target group member")
	}
	return nil
}

func (r *Repository) ListGroupMembers(ctx context.Context, groupID string) ([]GroupMember, error) {
	query := r.db.Rebind(`SELECT * FROM target_group_members
		WHERE group_id = ?
		ORDER BY added_at DESC, target_type, target_id`)
	ret := []GroupMember{}
	if err := r.db.SelectContext(ctx, &ret, query, strings.TrimSpace(groupID)); err != nil {
		return nil, errors.Wrap(err, "list target group members")
	}
	return ret, nil
}

func (r *Repository) CreateLog(ctx context.Context, input CreateLogInput) (*AnnotationLog, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("annotation repository database is nil")
	}
	if strings.TrimSpace(input.Title) == "" && strings.TrimSpace(input.BodyMarkdown) == "" {
		return nil, fmt.Errorf("title or body_markdown is required")
	}
	id := r.newID()
	query := r.db.Rebind(`INSERT INTO annotation_logs (
		id, log_kind, title, body_markdown, source_kind, source_label,
		agent_run_id, created_by, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	if _, err := r.db.ExecContext(
		ctx,
		query,
		id,
		defaultLogKind(input.LogKind),
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.BodyMarkdown),
		defaultSourceKind(input.SourceKind),
		strings.TrimSpace(input.SourceLabel),
		strings.TrimSpace(input.AgentRunID),
		strings.TrimSpace(input.CreatedBy),
	); err != nil {
		return nil, errors.Wrap(err, "insert annotation log")
	}
	return r.GetLog(ctx, id)
}

func (r *Repository) GetLog(ctx context.Context, id string) (*AnnotationLog, error) {
	var ret AnnotationLog
	query := r.db.Rebind(`SELECT * FROM annotation_logs WHERE id = ?`)
	if err := r.db.GetContext(ctx, &ret, query, id); err != nil {
		return nil, errors.Wrap(err, "get annotation log")
	}
	return &ret, nil
}

func (r *Repository) ListLogs(ctx context.Context, filter ListLogsFilter) ([]AnnotationLog, error) {
	query := `SELECT * FROM annotation_logs WHERE 1 = 1`
	args := make([]any, 0, 3)
	if strings.TrimSpace(filter.SourceKind) != "" {
		query += ` AND source_kind = ?`
		args = append(args, strings.TrimSpace(filter.SourceKind))
	}
	if strings.TrimSpace(filter.AgentRunID) != "" {
		query += ` AND agent_run_id = ?`
		args = append(args, strings.TrimSpace(filter.AgentRunID))
	}
	query += ` ORDER BY created_at DESC, id DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}
	ret := []AnnotationLog{}
	if err := r.db.SelectContext(ctx, &ret, r.db.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "list annotation logs")
	}
	return ret, nil
}

func (r *Repository) LinkLogTarget(ctx context.Context, input LinkLogTargetInput) error {
	if strings.TrimSpace(input.LogID) == "" {
		return fmt.Errorf("log_id is required")
	}
	if strings.TrimSpace(input.TargetType) == "" {
		return fmt.Errorf("target_type is required")
	}
	if strings.TrimSpace(input.TargetID) == "" {
		return fmt.Errorf("target_id is required")
	}
	query := r.db.Rebind(`INSERT INTO annotation_log_targets (
		log_id, target_type, target_id
	) VALUES (?, ?, ?)
	ON CONFLICT(log_id, target_type, target_id) DO NOTHING`)
	if _, err := r.db.ExecContext(
		ctx,
		query,
		strings.TrimSpace(input.LogID),
		strings.TrimSpace(input.TargetType),
		strings.TrimSpace(input.TargetID),
	); err != nil {
		return errors.Wrap(err, "insert annotation log target")
	}
	return nil
}

func (r *Repository) ListLogTargets(ctx context.Context, logID string) ([]LogTarget, error) {
	query := r.db.Rebind(`SELECT * FROM annotation_log_targets
		WHERE log_id = ?
		ORDER BY target_type, target_id`)
	ret := []LogTarget{}
	if err := r.db.SelectContext(ctx, &ret, query, strings.TrimSpace(logID)); err != nil {
		return nil, errors.Wrap(err, "list annotation log targets")
	}
	return ret, nil
}

func (r *Repository) ListRuns(ctx context.Context) ([]AgentRunSummary, error) {
	query := `
WITH run_annotations AS (
	SELECT
		agent_run_id,
		COUNT(*) AS annotation_count,
		SUM(CASE WHEN review_state = 'to_review' THEN 1 ELSE 0 END) AS pending_count,
		SUM(CASE WHEN review_state = 'reviewed' THEN 1 ELSE 0 END) AS reviewed_count,
		SUM(CASE WHEN review_state = 'dismissed' THEN 1 ELSE 0 END) AS dismissed_count,
		MIN(created_at) AS started_at,
		MAX(created_at) AS completed_at
	FROM annotations
	WHERE agent_run_id != ''
	GROUP BY agent_run_id
),
run_annotation_metadata AS (
	SELECT agent_run_id, source_label, source_kind
	FROM (
		SELECT
			agent_run_id,
			source_label,
			source_kind,
			ROW_NUMBER() OVER (
				PARTITION BY agent_run_id
				ORDER BY created_at DESC, id DESC
			) AS row_number
		FROM annotations
		WHERE agent_run_id != ''
	)
	WHERE row_number = 1
),
run_logs AS (
	SELECT agent_run_id, COUNT(*) AS log_count
	FROM annotation_logs
	WHERE agent_run_id != ''
	GROUP BY agent_run_id
),
run_groups AS (
	SELECT agent_run_id, COUNT(*) AS group_count
	FROM target_groups
	WHERE agent_run_id != ''
	GROUP BY agent_run_id
)
SELECT
	ra.agent_run_id AS run_id,
	COALESCE(ram.source_label, '') AS source_label,
	COALESCE(ram.source_kind, '') AS source_kind,
	ra.annotation_count,
	ra.pending_count,
	ra.reviewed_count,
	ra.dismissed_count,
	COALESCE(rl.log_count, 0) AS log_count,
	COALESCE(rg.group_count, 0) AS group_count,
	ra.started_at,
	ra.completed_at
FROM run_annotations ra
LEFT JOIN run_annotation_metadata ram ON ram.agent_run_id = ra.agent_run_id
LEFT JOIN run_logs rl ON rl.agent_run_id = ra.agent_run_id
LEFT JOIN run_groups rg ON rg.agent_run_id = ra.agent_run_id
ORDER BY ra.started_at DESC, ra.agent_run_id DESC`

	ret := []AgentRunSummary{}
	if err := r.db.SelectContext(ctx, &ret, query); err != nil {
		return nil, errors.Wrap(err, "list annotation runs")
	}
	return ret, nil
}

func (r *Repository) GetRunDetail(ctx context.Context, runID string) (*AgentRunDetail, error) {
	summary, err := r.getRunSummary(ctx, runID)
	if err != nil {
		return nil, err
	}

	annotations, err := r.ListAnnotations(ctx, ListAnnotationsFilter{AgentRunID: runID})
	if err != nil {
		return nil, errors.Wrap(err, "list run annotations")
	}

	logs, err := r.ListLogs(ctx, ListLogsFilter{AgentRunID: runID})
	if err != nil {
		return nil, errors.Wrap(err, "list run logs")
	}

	groups := []TargetGroup{}
	query := r.db.Rebind(`SELECT * FROM target_groups
		WHERE agent_run_id = ?
		ORDER BY created_at DESC, id DESC`)
	if err := r.db.SelectContext(ctx, &groups, query, strings.TrimSpace(runID)); err != nil {
		return nil, errors.Wrap(err, "list run groups")
	}

	return &AgentRunDetail{
		AgentRunSummary: *summary,
		Annotations:     annotations,
		Logs:            logs,
		Groups:          groups,
	}, nil
}

func (r *Repository) getRunSummary(ctx context.Context, runID string) (*AgentRunSummary, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, fmt.Errorf("run id is required")
	}

	query := `
WITH run_annotations AS (
	SELECT
		agent_run_id,
		COUNT(*) AS annotation_count,
		SUM(CASE WHEN review_state = 'to_review' THEN 1 ELSE 0 END) AS pending_count,
		SUM(CASE WHEN review_state = 'reviewed' THEN 1 ELSE 0 END) AS reviewed_count,
		SUM(CASE WHEN review_state = 'dismissed' THEN 1 ELSE 0 END) AS dismissed_count,
		MIN(created_at) AS started_at,
		MAX(created_at) AS completed_at
	FROM annotations
	WHERE agent_run_id = ?
	GROUP BY agent_run_id
),
run_annotation_metadata AS (
	SELECT agent_run_id, source_label, source_kind
	FROM (
		SELECT
			agent_run_id,
			source_label,
			source_kind,
			ROW_NUMBER() OVER (
				PARTITION BY agent_run_id
				ORDER BY created_at DESC, id DESC
			) AS row_number
		FROM annotations
		WHERE agent_run_id = ?
	)
	WHERE row_number = 1
),
run_logs AS (
	SELECT agent_run_id, COUNT(*) AS log_count
	FROM annotation_logs
	WHERE agent_run_id = ?
	GROUP BY agent_run_id
),
run_groups AS (
	SELECT agent_run_id, COUNT(*) AS group_count
	FROM target_groups
	WHERE agent_run_id = ?
	GROUP BY agent_run_id
)
SELECT
	ra.agent_run_id AS run_id,
	COALESCE(ram.source_label, '') AS source_label,
	COALESCE(ram.source_kind, '') AS source_kind,
	ra.annotation_count,
	ra.pending_count,
	ra.reviewed_count,
	ra.dismissed_count,
	COALESCE(rl.log_count, 0) AS log_count,
	COALESCE(rg.group_count, 0) AS group_count,
	ra.started_at,
	ra.completed_at
FROM run_annotations ra
LEFT JOIN run_annotation_metadata ram ON ram.agent_run_id = ra.agent_run_id
LEFT JOIN run_logs rl ON rl.agent_run_id = ra.agent_run_id
LEFT JOIN run_groups rg ON rg.agent_run_id = ra.agent_run_id`

	var ret AgentRunSummary
	if err := r.db.GetContext(ctx, &ret, r.db.Rebind(query), runID, runID, runID, runID); err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("annotation run %s not found", runID)
		}
		return nil, errors.Wrap(err, "get annotation run summary")
	}
	return &ret, nil
}

func defaultSourceKind(sourceKind string) string {
	sourceKind = strings.TrimSpace(sourceKind)
	if sourceKind == "" {
		return SourceKindHuman
	}
	return sourceKind
}

func defaultReviewState(reviewState, sourceKind string) string {
	reviewState = strings.TrimSpace(reviewState)
	if reviewState != "" {
		return reviewState
	}
	if strings.TrimSpace(sourceKind) == SourceKindAgent {
		return ReviewStateToReview
	}
	return ReviewStateReviewed
}

func defaultLogKind(logKind string) string {
	logKind = strings.TrimSpace(logKind)
	if logKind == "" {
		return "note"
	}
	return logKind
}
