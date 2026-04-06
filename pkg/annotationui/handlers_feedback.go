package annotationui

import (
	"net/http"
	"strings"

	"github.com/go-go-golems/smailnail/pkg/annotate"
)

// ── Review Feedback Handlers ──────────────────────────────────────

func (h *appHandler) handleListFeedback(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 200)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	feedback, err := h.annotations.ListReviewFeedback(r.Context(), annotate.ListFeedbackFilter{
		AgentRunID:   strings.TrimSpace(r.URL.Query().Get("agentRunId")),
		Status:       strings.TrimSpace(r.URL.Query().Get("status")),
		FeedbackKind: strings.TrimSpace(r.URL.Query().Get("feedbackKind")),
		MailboxName:  strings.TrimSpace(r.URL.Query().Get("mailboxName")),
		Limit:        limit,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ret := make([]feedbackResponse, 0, len(feedback))
	for i := range feedback {
		ret = append(ret, feedbackToResponse(&feedback[i]))
	}
	writeJSON(w, http.StatusOK, ret)
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
	writeJSON(w, http.StatusOK, feedbackToResponse(fb))
}

func (h *appHandler) handleCreateFeedback(w http.ResponseWriter, r *http.Request) {
	req := createFeedbackRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if strings.TrimSpace(req.FeedbackKind) == "" {
		req.FeedbackKind = annotate.FeedbackKindComment
	}
	if strings.TrimSpace(req.ScopeKind) == "" {
		req.ScopeKind = annotate.FeedbackScopeRun
	}

	targets := make([]annotate.FeedbackTargetInput, 0, len(req.Targets))
	for _, t := range req.Targets {
		targets = append(targets, annotate.FeedbackTargetInput{
			TargetType: strings.TrimSpace(t.TargetType),
			TargetID:   strings.TrimSpace(t.TargetID),
		})
	}

	fb, err := h.annotations.CreateReviewFeedback(r.Context(), annotate.CreateFeedbackInput{
		ScopeKind:    req.ScopeKind,
		AgentRunID:   req.AgentRunID,
		MailboxName:  req.MailboxName,
		FeedbackKind: req.FeedbackKind,
		Title:        req.Title,
		BodyMarkdown: req.BodyMarkdown,
		Targets:      targets,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, feedbackToResponse(fb))
}

func (h *appHandler) handleUpdateFeedback(w http.ResponseWriter, r *http.Request) {
	req := updateFeedbackRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if !isValidFeedbackStatus(req.Status) {
		writeMessageError(w, http.StatusBadRequest, "status must be one of open, acknowledged, resolved, archived")
		return
	}

	fb, err := h.annotations.UpdateReviewFeedback(r.Context(), r.PathValue("id"), annotate.UpdateFeedbackInput{
		Status: req.Status,
	})
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, feedbackToResponse(fb))
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

	ret := make([]guidelineResponse, 0, len(guidelines))
	for i := range guidelines {
		ret = append(ret, guidelineToResponse(&guidelines[i]))
	}
	writeJSON(w, http.StatusOK, ret)
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
	writeJSON(w, http.StatusOK, guidelineToResponse(g))
}

func (h *appHandler) handleCreateGuideline(w http.ResponseWriter, r *http.Request) {
	req := createGuidelineRequest{}
	if !decodeJSONBody(w, r, &req) {
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
		Slug:         req.Slug,
		Title:        req.Title,
		ScopeKind:    req.ScopeKind,
		BodyMarkdown: req.BodyMarkdown,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeMessageError(w, http.StatusConflict, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, guidelineToResponse(g))
}

func (h *appHandler) handleUpdateGuideline(w http.ResponseWriter, r *http.Request) {
	req := updateGuidelineRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if req.Status != nil && !isValidGuidelineStatus(*req.Status) {
		writeMessageError(w, http.StatusBadRequest, "status must be one of active, archived, draft")
		return
	}

	g, err := h.annotations.UpdateGuideline(r.Context(), r.PathValue("id"), annotate.UpdateGuidelineInput{
		Title:        req.Title,
		ScopeKind:    req.ScopeKind,
		Status:       req.Status,
		Priority:     req.Priority,
		BodyMarkdown: req.BodyMarkdown,
	})
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, guidelineToResponse(g))
}

// ── Run-Guideline Link Handlers ───────────────────────────────────

func (h *appHandler) handleListRunGuidelines(w http.ResponseWriter, r *http.Request) {
	guidelines, err := h.annotations.ListRunGuidelines(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ret := make([]guidelineResponse, 0, len(guidelines))
	for i := range guidelines {
		ret = append(ret, guidelineToResponse(&guidelines[i]))
	}
	writeJSON(w, http.StatusOK, ret)
}

func (h *appHandler) handleLinkRunGuideline(w http.ResponseWriter, r *http.Request) {
	req := linkGuidelineRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.GuidelineID) == "" {
		writeMessageError(w, http.StatusBadRequest, "guidelineId is required")
		return
	}

	if err := h.annotations.LinkGuidelineToRun(r.Context(),
		r.PathValue("id"),
		req.GuidelineID,
		"",
	); err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the updated list of linked guidelines
	guidelines, err := h.annotations.ListRunGuidelines(r.Context(), r.PathValue("id"))
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ret := make([]guidelineResponse, 0, len(guidelines))
	for i := range guidelines {
		ret = append(ret, guidelineToResponse(&guidelines[i]))
	}
	writeJSON(w, http.StatusOK, ret)
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

// ── Extend existing review handlers ───────────────────────────────
