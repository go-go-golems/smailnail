package annotate

// SchemaMigrationV3Statements returns the annotation MVP schema additions.
func SchemaMigrationV3Statements() []string {
	return SchemaMigrationV3CoreStatements()
}

func SchemaMigrationV3CoreStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS annotations (
			id TEXT PRIMARY KEY,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			tag TEXT NOT NULL DEFAULT '',
			note_markdown TEXT NOT NULL DEFAULT '',
			source_kind TEXT NOT NULL DEFAULT 'human',
			source_label TEXT NOT NULL DEFAULT '',
			agent_run_id TEXT NOT NULL DEFAULT '',
			review_state TEXT NOT NULL DEFAULT 'to_review',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_annotations_target
			ON annotations(target_type, target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_annotations_review_state
			ON annotations(review_state, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_annotations_tag
			ON annotations(tag, created_at)`,
		`CREATE TABLE IF NOT EXISTS target_groups (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			source_kind TEXT NOT NULL DEFAULT 'human',
			source_label TEXT NOT NULL DEFAULT '',
			agent_run_id TEXT NOT NULL DEFAULT '',
			review_state TEXT NOT NULL DEFAULT 'to_review',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_target_groups_review_state
			ON target_groups(review_state, created_at)`,
		`CREATE TABLE IF NOT EXISTS target_group_members (
			group_id TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, target_type, target_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_target_group_members_target
			ON target_group_members(target_type, target_id)`,
		`CREATE TABLE IF NOT EXISTS annotation_logs (
			id TEXT PRIMARY KEY,
			log_kind TEXT NOT NULL DEFAULT 'note',
			title TEXT NOT NULL DEFAULT '',
			body_markdown TEXT NOT NULL DEFAULT '',
			source_kind TEXT NOT NULL DEFAULT 'human',
			source_label TEXT NOT NULL DEFAULT '',
			agent_run_id TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_annotation_logs_created_at
			ON annotation_logs(created_at)`,
		`CREATE TABLE IF NOT EXISTS annotation_log_targets (
			log_id TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			PRIMARY KEY (log_id, target_type, target_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_annotation_log_targets_target
			ON annotation_log_targets(target_type, target_id)`,
	}
}

// SchemaMigrationV4Statements returns schema additions for review feedback,
// review guidelines, and run-guideline links.
func SchemaMigrationV4Statements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS review_feedback (
			id TEXT PRIMARY KEY,
			scope_kind TEXT NOT NULL DEFAULT 'selection',
			agent_run_id TEXT NOT NULL DEFAULT '',
			mailbox_name TEXT NOT NULL DEFAULT '',
			feedback_kind TEXT NOT NULL DEFAULT 'comment',
			status TEXT NOT NULL DEFAULT 'open',
			title TEXT NOT NULL DEFAULT '',
			body_markdown TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_review_feedback_run
			ON review_feedback(agent_run_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_review_feedback_status
			ON review_feedback(status, created_at)`,
		`CREATE TABLE IF NOT EXISTS review_feedback_targets (
			feedback_id TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			PRIMARY KEY (feedback_id, target_type, target_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_review_feedback_targets_target
			ON review_feedback_targets(target_type, target_id)`,
		`CREATE TABLE IF NOT EXISTS review_guidelines (
			id TEXT PRIMARY KEY,
			slug TEXT NOT NULL UNIQUE,
			title TEXT NOT NULL,
			scope_kind TEXT NOT NULL DEFAULT 'global',
			status TEXT NOT NULL DEFAULT 'active',
			priority INTEGER NOT NULL DEFAULT 0,
			body_markdown TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_review_guidelines_status
			ON review_guidelines(status, priority DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_review_guidelines_slug
			ON review_guidelines(slug)`,
		`CREATE TABLE IF NOT EXISTS run_guideline_links (
			agent_run_id TEXT NOT NULL,
			guideline_id TEXT NOT NULL,
			linked_by TEXT NOT NULL DEFAULT '',
			linked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (agent_run_id, guideline_id)
		)`,
	}
}
