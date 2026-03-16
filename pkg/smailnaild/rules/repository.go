package rules

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("rule not found")

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, record *RuleRecord) error {
	if record == nil {
		return fmt.Errorf("rule record is nil")
	}

	_, err := r.db.NamedExecContext(ctx, `INSERT INTO rules (
		id,
		user_id,
		imap_account_id,
		name,
		description,
		status,
		source_kind,
		rule_yaml,
		last_preview_count,
		last_run_at
	) VALUES (
		:id,
		:user_id,
		:imap_account_id,
		:name,
		:description,
		:status,
		:source_kind,
		:rule_yaml,
		:last_preview_count,
		:last_run_at
	)`, record)
	return err
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]RuleRecord, error) {
	var ret []RuleRecord
	err := r.db.SelectContext(ctx, &ret, r.db.Rebind(`SELECT
		id,
		user_id,
		imap_account_id,
		name,
		description,
		status,
		source_kind,
		rule_yaml,
		last_preview_count,
		last_run_at,
		created_at,
		updated_at
	FROM rules
	WHERE user_id = ?
	ORDER BY created_at ASC`), userID)
	return ret, err
}

func (r *Repository) GetByID(ctx context.Context, userID, ruleID string) (*RuleRecord, error) {
	var record RuleRecord
	err := r.db.GetContext(ctx, &record, r.db.Rebind(`SELECT
		id,
		user_id,
		imap_account_id,
		name,
		description,
		status,
		source_kind,
		rule_yaml,
		last_preview_count,
		last_run_at,
		created_at,
		updated_at
	FROM rules
	WHERE id = ? AND user_id = ?`), ruleID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *Repository) Update(ctx context.Context, record *RuleRecord) error {
	if record == nil {
		return fmt.Errorf("rule record is nil")
	}

	result, err := r.db.NamedExecContext(ctx, `UPDATE rules SET
		imap_account_id = :imap_account_id,
		name = :name,
		description = :description,
		status = :status,
		source_kind = :source_kind,
		rule_yaml = :rule_yaml,
		last_preview_count = :last_preview_count,
		last_run_at = :last_run_at,
		updated_at = CURRENT_TIMESTAMP
	WHERE id = :id AND user_id = :user_id`, record)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, ruleID string) error {
	if _, err := r.db.ExecContext(ctx, r.db.Rebind(`DELETE FROM rule_runs WHERE rule_id = ? AND user_id = ?`), ruleID, userID); err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, r.db.Rebind(`DELETE FROM rules WHERE id = ? AND user_id = ?`), ruleID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CreateRun(ctx context.Context, run *RuleRun) error {
	if run == nil {
		return fmt.Errorf("rule run is nil")
	}

	_, err := r.db.NamedExecContext(ctx, `INSERT INTO rule_runs (
		id,
		rule_id,
		user_id,
		imap_account_id,
		mode,
		matched_count,
		action_summary_json,
		sample_results_json
	) VALUES (
		:id,
		:rule_id,
		:user_id,
		:imap_account_id,
		:mode,
		:matched_count,
		:action_summary_json,
		:sample_results_json
	)`, run)
	return err
}
