package rules

import "time"

type RuleRecord struct {
	ID               string    `db:"id" json:"id"`
	UserID           string    `db:"user_id" json:"userId"`
	IMAPAccountID    string    `db:"imap_account_id" json:"imapAccountId"`
	Name             string    `db:"name" json:"name"`
	Description      string    `db:"description" json:"description"`
	Status           string    `db:"status" json:"status"`
	SourceKind       string    `db:"source_kind" json:"sourceKind"`
	RuleYAML         string    `db:"rule_yaml" json:"ruleYaml"`
	LastPreviewCount int       `db:"last_preview_count" json:"lastPreviewCount"`
	LastRunAt        time.Time `db:"last_run_at" json:"lastRunAt"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at" json:"updatedAt"`
}

type RuleRun struct {
	ID                string    `db:"id" json:"id"`
	RuleID            string    `db:"rule_id" json:"ruleId"`
	UserID            string    `db:"user_id" json:"userId"`
	IMAPAccountID     string    `db:"imap_account_id" json:"imapAccountId"`
	Mode              string    `db:"mode" json:"mode"`
	MatchedCount      int       `db:"matched_count" json:"matchedCount"`
	ActionSummaryJSON string    `db:"action_summary_json" json:"actionSummaryJson"`
	SampleResultsJSON string    `db:"sample_results_json" json:"sampleResultsJson"`
	CreatedAt         time.Time `db:"created_at" json:"createdAt"`
}
