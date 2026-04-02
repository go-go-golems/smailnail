package annotate

import "time"

const (
	SourceKindHuman     = "human"
	SourceKindAgent     = "agent"
	SourceKindHeuristic = "heuristic"
	SourceKindImport    = "import"

	ReviewStateToReview  = "to_review"
	ReviewStateReviewed  = "reviewed"
	ReviewStateDismissed = "dismissed"
)

type Annotation struct {
	ID           string    `db:"id" json:"id"`
	TargetType   string    `db:"target_type" json:"targetType"`
	TargetID     string    `db:"target_id" json:"targetId"`
	Tag          string    `db:"tag" json:"tag"`
	NoteMarkdown string    `db:"note_markdown" json:"noteMarkdown"`
	SourceKind   string    `db:"source_kind" json:"sourceKind"`
	SourceLabel  string    `db:"source_label" json:"sourceLabel"`
	AgentRunID   string    `db:"agent_run_id" json:"agentRunId"`
	ReviewState  string    `db:"review_state" json:"reviewState"`
	CreatedBy    string    `db:"created_by" json:"createdBy"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateAnnotationInput struct {
	TargetType   string
	TargetID     string
	Tag          string
	NoteMarkdown string
	SourceKind   string
	SourceLabel  string
	AgentRunID   string
	ReviewState  string
	CreatedBy    string
}

type ListAnnotationsFilter struct {
	TargetType  string
	TargetID    string
	Tag         string
	ReviewState string
	SourceKind  string
	Limit       int
}

type TargetGroup struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	SourceKind  string    `db:"source_kind" json:"sourceKind"`
	SourceLabel string    `db:"source_label" json:"sourceLabel"`
	AgentRunID  string    `db:"agent_run_id" json:"agentRunId"`
	ReviewState string    `db:"review_state" json:"reviewState"`
	CreatedBy   string    `db:"created_by" json:"createdBy"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateGroupInput struct {
	Name        string
	Description string
	SourceKind  string
	SourceLabel string
	AgentRunID  string
	ReviewState string
	CreatedBy   string
}

type ListGroupsFilter struct {
	ReviewState string
	SourceKind  string
	Limit       int
}

type GroupMember struct {
	GroupID    string    `db:"group_id" json:"groupId"`
	TargetType string    `db:"target_type" json:"targetType"`
	TargetID   string    `db:"target_id" json:"targetId"`
	AddedAt    time.Time `db:"added_at" json:"addedAt"`
}

type AddGroupMemberInput struct {
	GroupID    string
	TargetType string
	TargetID   string
}

type AnnotationLog struct {
	ID           string    `db:"id" json:"id"`
	LogKind      string    `db:"log_kind" json:"logKind"`
	Title        string    `db:"title" json:"title"`
	BodyMarkdown string    `db:"body_markdown" json:"bodyMarkdown"`
	SourceKind   string    `db:"source_kind" json:"sourceKind"`
	SourceLabel  string    `db:"source_label" json:"sourceLabel"`
	AgentRunID   string    `db:"agent_run_id" json:"agentRunId"`
	CreatedBy    string    `db:"created_by" json:"createdBy"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

type CreateLogInput struct {
	LogKind      string
	Title        string
	BodyMarkdown string
	SourceKind   string
	SourceLabel  string
	AgentRunID   string
	CreatedBy    string
}

type ListLogsFilter struct {
	SourceKind string
	AgentRunID string
	Limit      int
}

type LogTarget struct {
	LogID      string `db:"log_id" json:"logId"`
	TargetType string `db:"target_type" json:"targetType"`
	TargetID   string `db:"target_id" json:"targetId"`
}

type LinkLogTargetInput struct {
	LogID      string
	TargetType string
	TargetID   string
}
