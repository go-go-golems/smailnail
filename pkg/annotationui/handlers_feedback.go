package annotationui

import (
	"net/http"
	"strings"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	annotationuiv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1"
)

// ── Review Feedback Handlers ──────────────────────────────────────

func (h *appHandler) handleListFeedback(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 200)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	feedback, err := h.annotations.ListReviewFeedback(r.Context(), annotate.ListFeedbackFilter{
		ScopeKind:    strings.TrimSpace(r.URL.Query().Get("scopeKind")),
		AgentRunID:   strings.TrimSpace(r.URL.Query().Get("agentRunId")),
		Status:       strings.TrimSpace(r.URL.Query().Get("status")),
		FeedbackKind: strings.TrimSpace(r.URL.Query().Get("feedbackKind")),
		MailboxName:  strings.TrimSpace(r.URL.Query().Get("mailboxName")),
		TargetType:   strings.TrimSpace(r.URL.Query().Get("targetType")),
		TargetID:     strings.TrimSpace(r.URL.Query().Get("targetId")),
		Limit:        limit,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeProtoJSON(w, http.StatusOK, feedbackListToProto(feedback))
}

func (h *appHandler) handleGetFeedback(w http.ResponseWriter, r *http.Request) {
	fb, err := h.annotations.GetReviewFeedback(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, feedbackToProto(fb))
}

func (h *appHandler) handleCreateFeedback(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.CreateFeedbackRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}

	feedbackKind := strings.TrimSpace(req.FeedbackKind)
	if feedbackKind == "" {
		feedbackKind = annotate.FeedbackKindComment
	}
	scopeKind := strings.TrimSpace(req.ScopeKind)
	if scopeKind == "" {
		scopeKind = annotate.FeedbackScopeRun
	}

	fb, err := h.annotations.CreateReviewFeedback(r.Context(), annotate.CreateFeedbackInput{
		ScopeKind:    scopeKind,
		AgentRunID:   strings.TrimSpace(req.AgentRunId),
		MailboxName:  strings.TrimSpace(req.MailboxName),
		FeedbackKind: feedbackKind,
		Title:        strings.TrimSpace(req.Title),
		BodyMarkdown: strings.TrimSpace(req.BodyMarkdown),
		CreatedBy:    requestReviewActor(r),
		Targets:      protoTargetsToAnnotate(req.Targets),
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusCreated, feedbackToProto(fb))
}

func (h *appHandler) handleUpdateFeedback(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.UpdateFeedbackRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	if req.Status == nil {
		writeMessageError(w, http.StatusBadRequest, "status is required")
		return
	}
	if !isValidFeedbackStatus(req.GetStatus()) {
		writeMessageError(w, http.StatusBadRequest, "status must be one of open, acknowledged, resolved, archived")
		return
	}

	fb, err := h.annotations.UpdateReviewFeedback(r.Context(), r.PathValue("id"), annotate.UpdateFeedbackInput{
		Status: strings.TrimSpace(req.GetStatus()),
	})
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, feedbackToProto(fb))
}

// ── Guideline Handlers ────────────────────────────────────────────

func (h *appHandler) handleListGuidelines(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 200)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	guidelines, err := h.annotations.ListGuidelines(r.Context(), annotate.ListGuidelinesFilter{
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		ScopeKind: strings.TrimSpace(r.URL.Query().Get("scopeKind")),
		Search:    strings.TrimSpace(r.URL.Query().Get("search")),
		Limit:     limit,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeProtoJSON(w, http.StatusOK, guidelineListToProto(guidelines))
}

func (h *appHandler) handleGetGuideline(w http.ResponseWriter, r *http.Request) {
	g, err := h.annotations.GetGuideline(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, guidelineToProto(g))
}

func (h *appHandler) handleListGuidelineRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := h.annotations.ListGuidelineRuns(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, &annotationuiv1.AgentRunListResponse{Items: annotateRunsToProto(runs)})
}

func (h *appHandler) handleCreateGuideline(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.CreateGuidelineRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	if strings.TrimSpace(req.Slug) == "" {
		writeMessageError(w, http.StatusBadRequest, "slug is required")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeMessageError(w, http.StatusBadRequest, "title is required")
		return
	}

	g, err := h.annotations.CreateGuideline(r.Context(), annotate.CreateGuidelineInput{
		Slug:         strings.TrimSpace(req.Slug),
		Title:        strings.TrimSpace(req.Title),
		ScopeKind:    strings.TrimSpace(req.ScopeKind),
		BodyMarkdown: strings.TrimSpace(req.BodyMarkdown),
		CreatedBy:    requestReviewActor(r),
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeMessageError(w, http.StatusConflict, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusCreated, guidelineToProto(g))
}

func (h *appHandler) handleUpdateGuideline(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.UpdateGuidelineRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	if req.Status != nil && !isValidGuidelineStatus(req.GetStatus()) {
		writeMessageError(w, http.StatusBadRequest, "status must be one of active, archived, draft")
		return
	}

	g, err := h.annotations.UpdateGuideline(r.Context(), r.PathValue("id"), protoUpdateGuidelineToAnnotate(req))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, guidelineToProto(g))
}

// ── Run-Guideline Link Handlers ───────────────────────────────────

func (h *appHandler) handleListRunGuidelines(w http.ResponseWriter, r *http.Request) {
	guidelines, err := h.annotations.ListRunGuidelines(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeProtoJSON(w, http.StatusOK, guidelineListToProto(guidelines))
}

func (h *appHandler) handleLinkRunGuideline(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.LinkRunGuidelineRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	if strings.TrimSpace(req.GuidelineId) == "" {
		writeMessageError(w, http.StatusBadRequest, "guidelineId is required")
		return
	}

	if err := h.annotations.LinkGuidelineToRun(r.Context(),
		r.PathValue("id"),
		strings.TrimSpace(req.GuidelineId),
		requestReviewActor(r),
	); err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	guidelines, err := h.annotations.ListRunGuidelines(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, guidelineListToProto(guidelines))
}

func (h *appHandler) handleUnlinkRunGuideline(w http.ResponseWriter, r *http.Request) {
	if err := h.annotations.UnlinkGuidelineFromRun(r.Context(),
		r.PathValue("id"),
		r.PathValue("guidelineId"),
	); err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
