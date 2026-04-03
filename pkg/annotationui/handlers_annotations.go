package annotationui

import (
	"net/http"
	"strings"

	"github.com/go-go-golems/smailnail/pkg/annotate"
)

type reviewRequest struct {
	ReviewState string `json:"reviewState"`
}

type batchReviewRequest struct {
	IDs         []string `json:"ids"`
	ReviewState string   `json:"reviewState"`
}

func (h *appHandler) handleListAnnotations(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 500)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	annotations, err := h.annotations.ListAnnotations(r.Context(), annotate.ListAnnotationsFilter{
		TargetType:  strings.TrimSpace(r.URL.Query().Get("targetType")),
		TargetID:    strings.TrimSpace(r.URL.Query().Get("targetId")),
		Tag:         strings.TrimSpace(r.URL.Query().Get("tag")),
		ReviewState: strings.TrimSpace(r.URL.Query().Get("reviewState")),
		SourceKind:  strings.TrimSpace(r.URL.Query().Get("sourceKind")),
		AgentRunID:  strings.TrimSpace(r.URL.Query().Get("agentRunId")),
		Limit:       limit,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, annotations)
}

func (h *appHandler) handleGetAnnotation(w http.ResponseWriter, r *http.Request) {
	annotation, err := h.annotations.GetAnnotation(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, annotation)
}

func (h *appHandler) handleReviewAnnotation(w http.ResponseWriter, r *http.Request) {
	req := reviewRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if !isValidReviewState(req.ReviewState) {
		writeMessageError(w, http.StatusBadRequest, "reviewState must be one of to_review, reviewed, dismissed")
		return
	}

	annotation, err := h.annotations.UpdateAnnotationReviewState(r.Context(), r.PathValue("id"), req.ReviewState)
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, annotation)
}

func (h *appHandler) handleBatchReview(w http.ResponseWriter, r *http.Request) {
	req := batchReviewRequest{}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if len(req.IDs) == 0 {
		writeMessageError(w, http.StatusBadRequest, "ids must not be empty")
		return
	}
	if !isValidReviewState(req.ReviewState) {
		writeMessageError(w, http.StatusBadRequest, "reviewState must be one of to_review, reviewed, dismissed")
		return
	}
	if err := h.annotations.BatchUpdateReviewState(r.Context(), req.IDs, req.ReviewState); err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *appHandler) handleListGroups(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 100)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}
	groups, err := h.annotations.ListGroups(r.Context(), annotate.ListGroupsFilter{
		ReviewState: strings.TrimSpace(r.URL.Query().Get("reviewState")),
		SourceKind:  strings.TrimSpace(r.URL.Query().Get("sourceKind")),
		Limit:       limit,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (h *appHandler) handleGetGroup(w http.ResponseWriter, r *http.Request) {
	group, err := h.annotations.GetGroup(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	members, err := h.annotations.ListGroupMembers(r.Context(), group.ID)
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, annotate.GroupDetail{
		TargetGroup: *group,
		Members:     members,
	})
}

func (h *appHandler) handleListLogs(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 200)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}
	logs, err := h.annotations.ListLogs(r.Context(), annotate.ListLogsFilter{
		SourceKind: strings.TrimSpace(r.URL.Query().Get("sourceKind")),
		AgentRunID: strings.TrimSpace(r.URL.Query().Get("agentRunId")),
		Limit:      limit,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *appHandler) handleGetLog(w http.ResponseWriter, r *http.Request) {
	logEntry, err := h.annotations.GetLog(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, logEntry)
}

func (h *appHandler) handleListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := h.annotations.ListRuns(r.Context())
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (h *appHandler) handleGetRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.annotations.GetRunDetail(r.Context(), r.PathValue("id"))
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func isValidReviewState(reviewState string) bool {
	switch strings.TrimSpace(reviewState) {
	case annotate.ReviewStateToReview, annotate.ReviewStateReviewed, annotate.ReviewStateDismissed:
		return true
	default:
		return false
	}
}
