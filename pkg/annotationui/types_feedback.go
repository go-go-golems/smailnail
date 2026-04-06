package annotationui

import (
	"strings"

	"github.com/go-go-golems/smailnail/pkg/annotate"
)

// ── Request/response types ──────────────────────────────────────

type feedbackResponse struct {
	ID           string               `json:"id"`
	ScopeKind    string               `json:"scopeKind"`
	AgentRunID   string               `json:"agentRunId"`
	MailboxName  string               `json:"mailboxName"`
	FeedbackKind string               `json:"feedbackKind"`
	Status       string               `json:"status"`
	Title        string               `json:"title"`
	BodyMarkdown string               `json:"bodyMarkdown"`
	CreatedBy    string               `json:"createdBy"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
	Targets      []feedbackTargetJSON `json:"targets"`
}

type feedbackTargetJSON struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
}

type createFeedbackRequest struct {
	ScopeKind    string               `json:"scopeKind"`
	AgentRunID   string               `json:"agentRunId"`
	MailboxName  string               `json:"mailboxName"`
	FeedbackKind string               `json:"feedbackKind"`
	Title        string               `json:"title"`
	BodyMarkdown string               `json:"bodyMarkdown"`
	Targets      []feedbackTargetJSON `json:"targets"`
}

type updateFeedbackRequest struct {
	Status string `json:"status"`
}

type guidelineResponse struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	ScopeKind    string `json:"scopeKind"`
	Status       string `json:"status"`
	Priority     int    `json:"priority"`
	BodyMarkdown string `json:"bodyMarkdown"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type createGuidelineRequest struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	ScopeKind    string `json:"scopeKind"`
	BodyMarkdown string `json:"bodyMarkdown"`
}

type updateGuidelineRequest struct {
	Title        *string `json:"title"`
	ScopeKind    *string `json:"scopeKind"`
	Status       *string `json:"status"`
	Priority     *int    `json:"priority"`
	BodyMarkdown *string `json:"bodyMarkdown"`
}

type linkGuidelineRequest struct {
	GuidelineID string `json:"guidelineId"`
}

type reviewCommentInput struct {
	FeedbackKind string `json:"feedbackKind"`
	Title        string `json:"title"`
	BodyMarkdown string `json:"bodyMarkdown"`
}

// ── Conversion helpers ──────────────────────────────────────────

func feedbackToResponse(f *annotate.ReviewFeedback) feedbackResponse {
	targets := make([]feedbackTargetJSON, 0, len(f.Targets))
	for _, t := range f.Targets {
		targets = append(targets, feedbackTargetJSON{
			TargetType: t.TargetType,
			TargetID:   t.TargetID,
		})
	}
	return feedbackResponse{
		ID:           f.ID,
		ScopeKind:    f.ScopeKind,
		AgentRunID:   f.AgentRunID,
		MailboxName:  f.MailboxName,
		FeedbackKind: f.FeedbackKind,
		Status:       f.Status,
		Title:        f.Title,
		BodyMarkdown: f.BodyMarkdown,
		CreatedBy:    f.CreatedBy,
		CreatedAt:    f.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    f.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Targets:      targets,
	}
}

func guidelineToResponse(g *annotate.ReviewGuideline) guidelineResponse {
	return guidelineResponse{
		ID:           g.ID,
		Slug:         g.Slug,
		Title:        g.Title,
		ScopeKind:    g.ScopeKind,
		Status:       g.Status,
		Priority:     g.Priority,
		BodyMarkdown: g.BodyMarkdown,
		CreatedBy:    g.CreatedBy,
		CreatedAt:    g.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    g.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func isValidFeedbackStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case annotate.FeedbackStatusOpen,
		annotate.FeedbackStatusAcknowledged,
		annotate.FeedbackStatusResolved,
		annotate.FeedbackStatusArchived:
		return true
	default:
		return false
	}
}

func isValidGuidelineStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case annotate.GuidelineStatusActive,
		annotate.GuidelineStatusArchived,
		annotate.GuidelineStatusDraft:
		return true
	default:
		return false
	}
}
