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

// ── Review Feedback ────────────────────────────────────────────

const (
	FeedbackScopeAnnotation = "annotation"
	FeedbackScopeSelection  = "selection"
	FeedbackScopeRun        = "run"
	FeedbackScopeGuideline  = "guideline"

	FeedbackKindComment          = "comment"
	FeedbackKindRejectRequest    = "reject_request"
	FeedbackKindGuidelineRequest = "guideline_request"
	FeedbackKindClarification    = "clarification"

	FeedbackStatusOpen         = "open"
	FeedbackStatusAcknowledged = "acknowledged"
	FeedbackStatusResolved     = "resolved"
	FeedbackStatusArchived     = "archived"
)

type ReviewFeedback struct {
	ID           string           `db:"id" json:"id"`
	ScopeKind    string           `db:"scope_kind" json:"scopeKind"`
	AgentRunID   string           `db:"agent_run_id" json:"agentRunId"`
	MailboxName  string           `db:"mailbox_name" json:"mailboxName"`
	FeedbackKind string           `db:"feedback_kind" json:"feedbackKind"`
	Status       string           `db:"status" json:"status"`
	Title        string           `db:"title" json:"title"`
	BodyMarkdown string           `db:"body_markdown" json:"bodyMarkdown"`
	CreatedBy    string           `db:"created_by" json:"createdBy"`
	CreatedAt    time.Time        `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time        `db:"updated_at" json:"updatedAt"`
	Targets      []FeedbackTarget `db:"-" json:"targets"`
}

type FeedbackTarget struct {
	FeedbackID string `db:"feedback_id" json:"feedbackId"`
	TargetType string `db:"target_type" json:"targetType"`
	TargetID   string `db:"target_id" json:"targetId"`
}

type CreateFeedbackInput struct {
	ScopeKind    string
	AgentRunID   string
	MailboxName  string
	FeedbackKind string
	Title        string
	BodyMarkdown string
	CreatedBy    string
	Targets      []FeedbackTargetInput
}

type FeedbackTargetInput struct {
	TargetType string
	TargetID   string
}

type ListFeedbackFilter struct {
	ScopeKind    string
	AgentRunID   string
	Status       string
	FeedbackKind string
	MailboxName  string
	Limit        int
}

type UpdateFeedbackInput struct {
	Status string
}

type ReviewCommentInput struct {
	FeedbackKind string
	Title        string
	BodyMarkdown string
}

type ReviewAnnotationActionInput struct {
	AnnotationID string
	ReviewState  string
	MailboxName  string
	Comment      *ReviewCommentInput
	GuidelineIDs []string
	CreatedBy    string
}

type BatchReviewActionInput struct {
	IDs          []string
	ReviewState  string
	AgentRunID   string
	MailboxName  string
	Comment      *ReviewCommentInput
	GuidelineIDs []string
	CreatedBy    string
}

// ── Review Guidelines ──────────────────────────────────────────

const (
	GuidelineScopeGlobal   = "global"
	GuidelineScopeMailbox  = "mailbox"
	GuidelineScopeSender   = "sender"
	GuidelineScopeDomain   = "domain"
	GuidelineScopeWorkflow = "workflow"

	GuidelineStatusActive   = "active"
	GuidelineStatusArchived = "archived"
	GuidelineStatusDraft    = "draft"
)

type ReviewGuideline struct {
	ID           string    `db:"id" json:"id"`
	Slug         string    `db:"slug" json:"slug"`
	Title        string    `db:"title" json:"title"`
	ScopeKind    string    `db:"scope_kind" json:"scopeKind"`
	Status       string    `db:"status" json:"status"`
	Priority     int       `db:"priority" json:"priority"`
	BodyMarkdown string    `db:"body_markdown" json:"bodyMarkdown"`
	CreatedBy    string    `db:"created_by" json:"createdBy"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

type CreateGuidelineInput struct {
	Slug         string
	Title        string
	ScopeKind    string
	BodyMarkdown string
	CreatedBy    string
}

type UpdateGuidelineInput struct {
	Title        *string
	ScopeKind    *string
	Status       *string
	Priority     *int
	BodyMarkdown *string
}

type ListGuidelinesFilter struct {
	Status    string
	ScopeKind string
	Search    string
	Limit     int
}

// ── Run-Guideline Links ───────────────────────────────────────

type RunGuidelineLink struct {
	AgentRunID  string    `db:"agent_run_id" json:"agentRunId"`
	GuidelineID string    `db:"guideline_id" json:"guidelineId"`
	LinkedBy    string    `db:"linked_by" json:"linkedBy"`
	LinkedAt    time.Time `db:"linked_at" json:"linkedAt"`
}
