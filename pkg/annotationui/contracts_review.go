package annotationui

import (
	"strings"
	"time"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	annotationuiv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1"
)

const annotationUITimeLayout = "2006-01-02T15:04:05Z"

func feedbackListToProto(items []annotate.ReviewFeedback) *annotationuiv1.ReviewFeedbackListResponse {
	ret := &annotationuiv1.ReviewFeedbackListResponse{Items: make([]*annotationuiv1.ReviewFeedback, 0, len(items))}
	for i := range items {
		ret.Items = append(ret.Items, feedbackToProto(&items[i]))
	}
	return ret
}

func feedbackToProto(f *annotate.ReviewFeedback) *annotationuiv1.ReviewFeedback {
	if f == nil {
		return &annotationuiv1.ReviewFeedback{}
	}
	targets := make([]*annotationuiv1.FeedbackTarget, 0, len(f.Targets))
	for _, t := range f.Targets {
		targets = append(targets, &annotationuiv1.FeedbackTarget{
			TargetType: t.TargetType,
			TargetId:   t.TargetID,
		})
	}
	return &annotationuiv1.ReviewFeedback{
		Id:           f.ID,
		ScopeKind:    f.ScopeKind,
		AgentRunId:   f.AgentRunID,
		MailboxName:  f.MailboxName,
		FeedbackKind: f.FeedbackKind,
		Status:       f.Status,
		Title:        f.Title,
		BodyMarkdown: f.BodyMarkdown,
		CreatedBy:    f.CreatedBy,
		CreatedAt:    formatAnnotationUITimestamp(f.CreatedAt),
		UpdatedAt:    formatAnnotationUITimestamp(f.UpdatedAt),
		Targets:      targets,
	}
}

func guidelineListToProto(items []annotate.ReviewGuideline) *annotationuiv1.ReviewGuidelineListResponse {
	ret := &annotationuiv1.ReviewGuidelineListResponse{Items: make([]*annotationuiv1.ReviewGuideline, 0, len(items))}
	for i := range items {
		ret.Items = append(ret.Items, guidelineToProto(&items[i]))
	}
	return ret
}

func guidelineToProto(g *annotate.ReviewGuideline) *annotationuiv1.ReviewGuideline {
	if g == nil {
		return &annotationuiv1.ReviewGuideline{}
	}
	return &annotationuiv1.ReviewGuideline{
		Id:           g.ID,
		Slug:         g.Slug,
		Title:        g.Title,
		ScopeKind:    g.ScopeKind,
		Status:       g.Status,
		Priority:     int32(g.Priority),
		BodyMarkdown: g.BodyMarkdown,
		CreatedBy:    g.CreatedBy,
		CreatedAt:    formatAnnotationUITimestamp(g.CreatedAt),
		UpdatedAt:    formatAnnotationUITimestamp(g.UpdatedAt),
	}
}

func protoTargetsToAnnotate(items []*annotationuiv1.FeedbackTarget) []annotate.FeedbackTargetInput {
	ret := make([]annotate.FeedbackTargetInput, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ret = append(ret, annotate.FeedbackTargetInput{
			TargetType: strings.TrimSpace(item.TargetType),
			TargetID:   strings.TrimSpace(item.TargetId),
		})
	}
	return ret
}

func protoCommentToAnnotate(comment *annotationuiv1.ReviewComment) *annotate.ReviewCommentInput {
	if comment == nil {
		return nil
	}
	if strings.TrimSpace(comment.BodyMarkdown) == "" && strings.TrimSpace(comment.Title) == "" {
		return nil
	}
	return &annotate.ReviewCommentInput{
		FeedbackKind: strings.TrimSpace(comment.FeedbackKind),
		Title:        strings.TrimSpace(comment.Title),
		BodyMarkdown: strings.TrimSpace(comment.BodyMarkdown),
	}
}

func protoUpdateGuidelineToAnnotate(req *annotationuiv1.UpdateGuidelineRequest) annotate.UpdateGuidelineInput {
	if req == nil {
		return annotate.UpdateGuidelineInput{}
	}
	ret := annotate.UpdateGuidelineInput{}
	if req.Title != nil {
		value := strings.TrimSpace(req.GetTitle())
		ret.Title = &value
	}
	if req.ScopeKind != nil {
		value := strings.TrimSpace(req.GetScopeKind())
		ret.ScopeKind = &value
	}
	if req.Status != nil {
		value := strings.TrimSpace(req.GetStatus())
		ret.Status = &value
	}
	if req.Priority != nil {
		value := int(req.GetPriority())
		ret.Priority = &value
	}
	if req.BodyMarkdown != nil {
		value := strings.TrimSpace(req.GetBodyMarkdown())
		ret.BodyMarkdown = &value
	}
	return ret
}

func formatAnnotationUITimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(annotationUITimeLayout)
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
