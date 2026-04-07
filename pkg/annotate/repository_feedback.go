package annotate

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ── Review Feedback ──────────────────────────────────────────────

func (r *Repository) CreateReviewFeedback(ctx context.Context, input CreateFeedbackInput) (*ReviewFeedback, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("annotation repository database is nil")
	}

	id := r.newID()
	feedbackKind := defaultString(input.FeedbackKind, FeedbackKindComment)
	scopeKind := defaultString(input.ScopeKind, FeedbackScopeSelection)

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin create feedback transaction")
	}
	defer func() { _ = tx.Rollback() }()

	query := tx.Rebind(`INSERT INTO review_feedback (
		id, scope_kind, agent_run_id, mailbox_name, feedback_kind, status,
		title, body_markdown, created_by,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, 'open', ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if _, err := tx.ExecContext(ctx, query,
		id, scopeKind,
		strings.TrimSpace(input.AgentRunID),
		strings.TrimSpace(input.MailboxName),
		feedbackKind,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.BodyMarkdown),
		strings.TrimSpace(input.CreatedBy),
	); err != nil {
		return nil, errors.Wrap(err, "insert review feedback")
	}

	// Insert targets
	for _, target := range input.Targets {
		tq := tx.Rebind(`INSERT INTO review_feedback_targets (
			feedback_id, target_type, target_id
		) VALUES (?, ?, ?)
		ON CONFLICT(feedback_id, target_type, target_id) DO NOTHING`)
		if _, err := tx.ExecContext(ctx, tq,
			id,
			strings.TrimSpace(target.TargetType),
			strings.TrimSpace(target.TargetID),
		); err != nil {
			return nil, errors.Wrap(err, "insert feedback target")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit create feedback")
	}

	return r.GetReviewFeedback(ctx, id)
}

func (r *Repository) GetReviewFeedback(ctx context.Context, id string) (*ReviewFeedback, error) {
	var feedback ReviewFeedback
	query := r.db.Rebind(`SELECT * FROM review_feedback WHERE id = ?`)
	if err := r.db.GetContext(ctx, &feedback, query, id); err != nil {
		return nil, errors.Wrap(err, "get review feedback")
	}

	targets, err := r.listFeedbackTargets(ctx, id)
	if err != nil {
		return nil, err
	}
	feedback.Targets = targets
	return &feedback, nil
}

func (r *Repository) ListReviewFeedback(ctx context.Context, filter ListFeedbackFilter) ([]ReviewFeedback, error) {
	query := `SELECT * FROM review_feedback WHERE 1 = 1`
	args := make([]any, 0, 5)

	if strings.TrimSpace(filter.AgentRunID) != "" {
		query += ` AND agent_run_id = ?`
		args = append(args, strings.TrimSpace(filter.AgentRunID))
	}
	if strings.TrimSpace(filter.Status) != "" {
		query += ` AND status = ?`
		args = append(args, strings.TrimSpace(filter.Status))
	}
	if strings.TrimSpace(filter.FeedbackKind) != "" {
		query += ` AND feedback_kind = ?`
		args = append(args, strings.TrimSpace(filter.FeedbackKind))
	}
	if strings.TrimSpace(filter.MailboxName) != "" {
		query += ` AND mailbox_name = ?`
		args = append(args, strings.TrimSpace(filter.MailboxName))
	}

	query += ` ORDER BY created_at DESC, id DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	var feedbacks []ReviewFeedback
	if err := r.db.SelectContext(ctx, &feedbacks, r.db.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "list review feedback")
	}

	// Attach targets for each feedback
	for i := range feedbacks {
		targets, err := r.listFeedbackTargets(ctx, feedbacks[i].ID)
		if err != nil {
			return nil, err
		}
		feedbacks[i].Targets = targets
	}

	return feedbacks, nil
}

func (r *Repository) UpdateReviewFeedback(ctx context.Context, id string, input UpdateFeedbackInput) (*ReviewFeedback, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}

	parts := []string{}
	args := make([]any, 0, 4)

	if strings.TrimSpace(input.Status) != "" {
		parts = append(parts, `status = ?`)
		args = append(args, strings.TrimSpace(input.Status))
	}

	if len(parts) == 0 {
		return r.GetReviewFeedback(ctx, id)
	}

	parts = append(parts, `updated_at = CURRENT_TIMESTAMP`)
	args = append(args, id)

	query := `UPDATE review_feedback SET ` + strings.Join(parts, ", ") + ` WHERE id = ?`
	result, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		return nil, errors.Wrap(err, "update review feedback")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("review feedback %s not found", id)
	}

	return r.GetReviewFeedback(ctx, id)
}

func (r *Repository) listFeedbackTargets(ctx context.Context, feedbackID string) ([]FeedbackTarget, error) {
	query := r.db.Rebind(`SELECT feedback_id, target_type, target_id
		FROM review_feedback_targets
		WHERE feedback_id = ?
		ORDER BY target_type, target_id`)
	var targets []FeedbackTarget
	if err := r.db.SelectContext(ctx, &targets, query, feedbackID); err != nil {
		return nil, errors.Wrap(err, "list feedback targets")
	}
	return targets, nil
}

// ── Review Guidelines ────────────────────────────────────────────

func (r *Repository) CreateGuideline(ctx context.Context, input CreateGuidelineInput) (*ReviewGuideline, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("annotation repository database is nil")
	}
	if strings.TrimSpace(input.Slug) == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}

	id := r.newID()
	scopeKind := defaultString(input.ScopeKind, GuidelineScopeGlobal)

	query := r.db.Rebind(`INSERT INTO review_guidelines (
		id, slug, title, scope_kind, status, priority, body_markdown, created_by,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, 'active', 0, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if _, err := r.db.ExecContext(ctx, query,
		id,
		strings.TrimSpace(input.Slug),
		strings.TrimSpace(input.Title),
		scopeKind,
		strings.TrimSpace(input.BodyMarkdown),
		strings.TrimSpace(input.CreatedBy),
	); err != nil {
		if isDuplicateKeyError(err) {
			return nil, fmt.Errorf("guideline with slug %q already exists", input.Slug)
		}
		return nil, errors.Wrap(err, "insert review guideline")
	}

	return r.GetGuideline(ctx, id)
}

func (r *Repository) GetGuideline(ctx context.Context, id string) (*ReviewGuideline, error) {
	var guideline ReviewGuideline
	query := r.db.Rebind(`SELECT * FROM review_guidelines WHERE id = ?`)
	if err := r.db.GetContext(ctx, &guideline, query, id); err != nil {
		return nil, errors.Wrap(err, "get review guideline")
	}
	return &guideline, nil
}

func (r *Repository) ListGuidelines(ctx context.Context, filter ListGuidelinesFilter) ([]ReviewGuideline, error) {
	query := `SELECT * FROM review_guidelines WHERE 1 = 1`
	args := make([]any, 0, 4)

	if strings.TrimSpace(filter.Status) != "" {
		query += ` AND status = ?`
		args = append(args, strings.TrimSpace(filter.Status))
	}
	if strings.TrimSpace(filter.ScopeKind) != "" {
		query += ` AND scope_kind = ?`
		args = append(args, strings.TrimSpace(filter.ScopeKind))
	}
	if strings.TrimSpace(filter.Search) != "" {
		query += ` AND (title LIKE ? OR slug LIKE ? OR body_markdown LIKE ?)`
		pattern := "%" + strings.TrimSpace(filter.Search) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	query += ` ORDER BY priority DESC, created_at DESC, id DESC`
	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	var guidelines []ReviewGuideline
	if err := r.db.SelectContext(ctx, &guidelines, r.db.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "list review guidelines")
	}
	return guidelines, nil
}

func (r *Repository) UpdateGuideline(ctx context.Context, id string, input UpdateGuidelineInput) (*ReviewGuideline, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}

	parts := []string{}
	args := make([]any, 0, 6)

	if input.Title != nil {
		parts = append(parts, `title = ?`)
		args = append(args, strings.TrimSpace(*input.Title))
	}
	if input.ScopeKind != nil {
		parts = append(parts, `scope_kind = ?`)
		args = append(args, strings.TrimSpace(*input.ScopeKind))
	}
	if input.Status != nil {
		parts = append(parts, `status = ?`)
		args = append(args, strings.TrimSpace(*input.Status))
	}
	if input.Priority != nil {
		parts = append(parts, `priority = ?`)
		args = append(args, *input.Priority)
	}
	if input.BodyMarkdown != nil {
		parts = append(parts, `body_markdown = ?`)
		args = append(args, strings.TrimSpace(*input.BodyMarkdown))
	}

	if len(parts) == 0 {
		return r.GetGuideline(ctx, id)
	}

	parts = append(parts, `updated_at = CURRENT_TIMESTAMP`)
	args = append(args, id)

	query := `UPDATE review_guidelines SET ` + strings.Join(parts, ", ") + ` WHERE id = ?`
	result, err := r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	if err != nil {
		return nil, errors.Wrap(err, "update review guideline")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("guideline %s not found", id)
	}

	return r.GetGuideline(ctx, id)
}

// ── Run-Guideline Links ──────────────────────────────────────────

func (r *Repository) LinkGuidelineToRun(ctx context.Context, runID, guidelineID, linkedBy string) error {
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(guidelineID) == "" {
		return fmt.Errorf("guideline_id is required")
	}
	query := r.db.Rebind(`INSERT INTO run_guideline_links (
		agent_run_id, guideline_id, linked_by, linked_at
	) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(agent_run_id, guideline_id) DO NOTHING`)
	if _, err := r.db.ExecContext(ctx, query,
		strings.TrimSpace(runID),
		strings.TrimSpace(guidelineID),
		strings.TrimSpace(linkedBy),
	); err != nil {
		return errors.Wrap(err, "link guideline to run")
	}
	return nil
}

func (r *Repository) UnlinkGuidelineFromRun(ctx context.Context, runID, guidelineID string) error {
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("run_id is required")
	}
	if strings.TrimSpace(guidelineID) == "" {
		return fmt.Errorf("guideline_id is required")
	}
	query := r.db.Rebind(`DELETE FROM run_guideline_links
		WHERE agent_run_id = ? AND guideline_id = ?`)
	_, err := r.db.ExecContext(ctx, query,
		strings.TrimSpace(runID),
		strings.TrimSpace(guidelineID),
	)
	if err != nil {
		return errors.Wrap(err, "unlink guideline from run")
	}
	return nil
}

func (r *Repository) ListRunGuidelines(ctx context.Context, runID string) ([]ReviewGuideline, error) {
	query := `
	SELECT g.*
	FROM review_guidelines g
	INNER JOIN run_guideline_links l ON l.guideline_id = g.id
	WHERE l.agent_run_id = ?
	ORDER BY g.priority DESC, g.created_at DESC, g.id DESC`

	var guidelines []ReviewGuideline
	if err := r.db.SelectContext(ctx, &guidelines, r.db.Rebind(query), strings.TrimSpace(runID)); err != nil {
		return nil, errors.Wrap(err, "list run guidelines")
	}
	return guidelines, nil
}

func (r *Repository) ListGuidelineRuns(ctx context.Context, guidelineID string) ([]AgentRunSummary, error) {
	query := `
WITH linked_runs AS (
	SELECT DISTINCT agent_run_id
	FROM run_guideline_links
	WHERE guideline_id = ? AND agent_run_id != ''
),
run_annotations AS (
	SELECT
		a.agent_run_id,
		COUNT(*) AS annotation_count,
		SUM(CASE WHEN a.review_state = 'to_review' THEN 1 ELSE 0 END) AS pending_count,
		SUM(CASE WHEN a.review_state = 'reviewed' THEN 1 ELSE 0 END) AS reviewed_count,
		SUM(CASE WHEN a.review_state = 'dismissed' THEN 1 ELSE 0 END) AS dismissed_count,
		MIN(a.created_at) AS started_at,
		MAX(a.created_at) AS completed_at
	FROM annotations a
	INNER JOIN linked_runs lr ON lr.agent_run_id = a.agent_run_id
	GROUP BY a.agent_run_id
),
run_annotation_metadata AS (
	SELECT agent_run_id, source_label, source_kind
	FROM (
		SELECT
			a.agent_run_id,
			a.source_label,
			a.source_kind,
			ROW_NUMBER() OVER (
				PARTITION BY a.agent_run_id
				ORDER BY a.created_at DESC, a.id DESC
			) AS row_number
		FROM annotations a
		INNER JOIN linked_runs lr ON lr.agent_run_id = a.agent_run_id
	)
	WHERE row_number = 1
),
run_logs AS (
	SELECT l.agent_run_id, COUNT(*) AS log_count
	FROM annotation_logs l
	INNER JOIN linked_runs lr ON lr.agent_run_id = l.agent_run_id
	GROUP BY l.agent_run_id
),
run_groups AS (
	SELECT g.agent_run_id, COUNT(*) AS group_count
	FROM target_groups g
	INNER JOIN linked_runs lr ON lr.agent_run_id = g.agent_run_id
	GROUP BY g.agent_run_id
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

	runs := []AgentRunSummary{}
	if err := r.db.SelectContext(ctx, &runs, r.db.Rebind(query), strings.TrimSpace(guidelineID)); err != nil {
		return nil, errors.Wrap(err, "list guideline runs")
	}
	return runs, nil
}

func (r *Repository) ReviewAnnotationWithArtifacts(ctx context.Context, input ReviewAnnotationActionInput) (*Annotation, error) {
	if strings.TrimSpace(input.AnnotationID) == "" {
		return nil, fmt.Errorf("annotation id is required")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin review annotation transaction")
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.updateAnnotationReviewStateTx(ctx, tx, input.AnnotationID, input.ReviewState); err != nil {
		return nil, err
	}

	annotation, err := r.getAnnotationTx(ctx, tx, input.AnnotationID)
	if err != nil {
		return nil, err
	}

	if input.Comment != nil && strings.TrimSpace(input.Comment.BodyMarkdown) != "" {
		if _, err := r.createReviewFeedbackTx(ctx, tx, CreateFeedbackInput{
			ScopeKind:    FeedbackScopeAnnotation,
			AgentRunID:   annotation.AgentRunID,
			MailboxName:  input.MailboxName,
			FeedbackKind: defaultString(input.Comment.FeedbackKind, FeedbackKindComment),
			Title:        input.Comment.Title,
			BodyMarkdown: input.Comment.BodyMarkdown,
			CreatedBy:    input.CreatedBy,
			Targets: []FeedbackTargetInput{{
				TargetType: FeedbackScopeAnnotation,
				TargetID:   annotation.ID,
			}},
		}); err != nil {
			return nil, err
		}
	}

	for _, guidelineID := range input.GuidelineIDs {
		guidelineID = strings.TrimSpace(guidelineID)
		if guidelineID == "" {
			continue
		}
		if strings.TrimSpace(annotation.AgentRunID) == "" {
			return nil, fmt.Errorf("cannot link guideline %s: annotation has no agent run id", guidelineID)
		}
		if err := r.linkGuidelineToRunTx(ctx, tx, annotation.AgentRunID, guidelineID, input.CreatedBy); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit review annotation transaction")
	}
	return r.GetAnnotation(ctx, input.AnnotationID)
}

func (r *Repository) BatchReviewWithArtifacts(ctx context.Context, input BatchReviewActionInput) error {
	if len(input.IDs) == 0 {
		return fmt.Errorf("ids are required")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin batch review with artifacts transaction")
	}
	defer func() { _ = tx.Rollback() }()

	trimmedIDs, err := r.batchUpdateReviewStateTx(ctx, tx, input.IDs, input.ReviewState)
	if err != nil {
		return err
	}

	runID := strings.TrimSpace(input.AgentRunID)
	if runID == "" && len(input.GuidelineIDs) > 0 {
		runID, err = r.inferSingleRunIDForAnnotationsTx(ctx, tx, trimmedIDs)
		if err != nil {
			return err
		}
	}

	if input.Comment != nil && strings.TrimSpace(input.Comment.BodyMarkdown) != "" {
		targets := make([]FeedbackTargetInput, 0, len(trimmedIDs))
		for _, id := range trimmedIDs {
			targets = append(targets, FeedbackTargetInput{TargetType: FeedbackScopeAnnotation, TargetID: id})
		}
		if _, err := r.createReviewFeedbackTx(ctx, tx, CreateFeedbackInput{
			ScopeKind:    FeedbackScopeSelection,
			AgentRunID:   runID,
			MailboxName:  input.MailboxName,
			FeedbackKind: defaultString(input.Comment.FeedbackKind, FeedbackKindComment),
			Title:        input.Comment.Title,
			BodyMarkdown: input.Comment.BodyMarkdown,
			CreatedBy:    input.CreatedBy,
			Targets:      targets,
		}); err != nil {
			return err
		}
	}

	for _, guidelineID := range input.GuidelineIDs {
		guidelineID = strings.TrimSpace(guidelineID)
		if guidelineID == "" {
			continue
		}
		if runID == "" {
			return fmt.Errorf("guideline linking requires agentRunId or a single run inferred from annotation ids")
		}
		if err := r.linkGuidelineToRunTx(ctx, tx, runID, guidelineID, input.CreatedBy); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit batch review with artifacts transaction")
	}
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────

func (r *Repository) getAnnotationTx(ctx context.Context, tx *sqlx.Tx, id string) (*Annotation, error) {
	var ret Annotation
	query := tx.Rebind(`SELECT * FROM annotations WHERE id = ?`)
	if err := tx.GetContext(ctx, &ret, query, id); err != nil {
		return nil, errors.Wrap(err, "get annotation in transaction")
	}
	return &ret, nil
}

func (r *Repository) updateAnnotationReviewStateTx(ctx context.Context, tx *sqlx.Tx, id, reviewState string) error {
	query := tx.Rebind(`UPDATE annotations
		SET review_state = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`)
	result, err := tx.ExecContext(ctx, query, defaultReviewState(reviewState, ""), strings.TrimSpace(id))
	if err != nil {
		return errors.Wrap(err, "update annotation review state in transaction")
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "read annotation review update result in transaction")
	}
	if rows == 0 {
		return fmt.Errorf("annotation %s not found", id)
	}
	return nil
}

func (r *Repository) batchUpdateReviewStateTx(ctx context.Context, tx *sqlx.Tx, ids []string, reviewState string) ([]string, error) {
	trimmedIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		trimmedIDs = append(trimmedIDs, id)
	}
	if len(trimmedIDs) == 0 {
		return nil, fmt.Errorf("ids are required")
	}

	query, args, err := sqlx.In(`UPDATE annotations
		SET review_state = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id IN (?)`, defaultReviewState(reviewState, ""), trimmedIDs)
	if err != nil {
		return nil, errors.Wrap(err, "build batch review update query in transaction")
	}
	if _, err := tx.ExecContext(ctx, tx.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "batch update annotation review state in transaction")
	}
	return trimmedIDs, nil
}

func (r *Repository) createReviewFeedbackTx(ctx context.Context, tx *sqlx.Tx, input CreateFeedbackInput) (string, error) {
	id := r.newID()
	feedbackKind := defaultString(input.FeedbackKind, FeedbackKindComment)
	scopeKind := defaultString(input.ScopeKind, FeedbackScopeSelection)

	query := tx.Rebind(`INSERT INTO review_feedback (
		id, scope_kind, agent_run_id, mailbox_name, feedback_kind, status,
		title, body_markdown, created_by,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, 'open', ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if _, err := tx.ExecContext(ctx, query,
		id, scopeKind,
		strings.TrimSpace(input.AgentRunID),
		strings.TrimSpace(input.MailboxName),
		feedbackKind,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.BodyMarkdown),
		strings.TrimSpace(input.CreatedBy),
	); err != nil {
		return "", errors.Wrap(err, "insert review feedback in transaction")
	}

	for _, target := range input.Targets {
		tq := tx.Rebind(`INSERT INTO review_feedback_targets (
			feedback_id, target_type, target_id
		) VALUES (?, ?, ?)
		ON CONFLICT(feedback_id, target_type, target_id) DO NOTHING`)
		if _, err := tx.ExecContext(ctx, tq,
			id,
			strings.TrimSpace(target.TargetType),
			strings.TrimSpace(target.TargetID),
		); err != nil {
			return "", errors.Wrap(err, "insert feedback target in transaction")
		}
	}

	return id, nil
}

func (r *Repository) linkGuidelineToRunTx(ctx context.Context, tx *sqlx.Tx, runID, guidelineID, linkedBy string) error {
	query := tx.Rebind(`INSERT INTO run_guideline_links (
		agent_run_id, guideline_id, linked_by, linked_at
	) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(agent_run_id, guideline_id) DO NOTHING`)
	if _, err := tx.ExecContext(ctx, query,
		strings.TrimSpace(runID),
		strings.TrimSpace(guidelineID),
		strings.TrimSpace(linkedBy),
	); err != nil {
		return errors.Wrap(err, "link guideline to run in transaction")
	}
	return nil
}

func (r *Repository) inferSingleRunIDForAnnotationsTx(ctx context.Context, tx *sqlx.Tx, ids []string) (string, error) {
	query, args, err := sqlx.In(`SELECT DISTINCT agent_run_id FROM annotations WHERE id IN (?)`, ids)
	if err != nil {
		return "", errors.Wrap(err, "build infer run id query")
	}
	var runIDs []string
	if err := tx.SelectContext(ctx, &runIDs, tx.Rebind(query), args...); err != nil {
		return "", errors.Wrap(err, "infer run id from annotations")
	}
	trimmed := make([]string, 0, len(runIDs))
	for _, runID := range runIDs {
		runID = strings.TrimSpace(runID)
		if runID == "" {
			continue
		}
		trimmed = append(trimmed, runID)
	}
	if len(trimmed) == 0 {
		return "", nil
	}
	if len(trimmed) > 1 {
		return "", fmt.Errorf("guideline linking requires annotations from a single run")
	}
	return trimmed[0], nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key")
}
