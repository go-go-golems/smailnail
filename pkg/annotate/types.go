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
	AgentRunID  string
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

type GroupDetail struct {
	TargetGroup
	Members []GroupMember `json:"members"`
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

type AgentRunSummary struct {
	RunID           string `db:"run_id" json:"runId"`
	SourceLabel     string `db:"source_label" json:"sourceLabel"`
	SourceKind      string `db:"source_kind" json:"sourceKind"`
	AnnotationCount int    `db:"annotation_count" json:"annotationCount"`
	PendingCount    int    `db:"pending_count" json:"pendingCount"`
	ReviewedCount   int    `db:"reviewed_count" json:"reviewedCount"`
	DismissedCount  int    `db:"dismissed_count" json:"dismissedCount"`
	LogCount        int    `db:"log_count" json:"logCount"`
	GroupCount      int    `db:"group_count" json:"groupCount"`
	StartedAt       string `db:"started_at" json:"startedAt"`
	CompletedAt     string `db:"completed_at" json:"completedAt"`
}

type AgentRunDetail struct {
	AgentRunSummary
	Annotations []Annotation    `json:"annotations"`
	Logs        []AnnotationLog `json:"logs"`
	Groups      []TargetGroup   `json:"groups"`
}
