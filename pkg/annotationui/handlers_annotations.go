package annotationui

import (
	"net/http"
	"strings"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	annotationuiv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1"
)

func (h *appHandler) handleListAnnotations(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 500)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	req := &annotationuiv1.AnnotationListRequest{
		TargetType:  strings.TrimSpace(r.URL.Query().Get("targetType")),
		TargetId:    strings.TrimSpace(r.URL.Query().Get("targetId")),
		Tag:         strings.TrimSpace(r.URL.Query().Get("tag")),
		ReviewState: strings.TrimSpace(r.URL.Query().Get("reviewState")),
		SourceKind:  strings.TrimSpace(r.URL.Query().Get("sourceKind")),
		AgentRunId:  strings.TrimSpace(r.URL.Query().Get("agentRunId")),
		Limit:       int32(limit),
	}

	annotations, err := h.annotations.ListAnnotations(r.Context(), annotate.ListAnnotationsFilter{
		TargetType:  req.GetTargetType(),
		TargetID:    req.GetTargetId(),
		Tag:         req.GetTag(),
		ReviewState: req.GetReviewState(),
		SourceKind:  req.GetSourceKind(),
		AgentRunID:  req.GetAgentRunId(),
		Limit:       int(req.GetLimit()),
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, &annotationuiv1.AnnotationListResponse{Items: annotateAnnotationsToProto(annotations)})
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
	writeProtoJSON(w, http.StatusOK, annotateAnnotationToProto(annotation))
}

func (h *appHandler) handleReviewAnnotation(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.ReviewAnnotationRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	if !isValidReviewState(req.GetReviewState()) {
		writeMessageError(w, http.StatusBadRequest, "reviewState must be one of to_review, reviewed, dismissed")
		return
	}

	annotation, err := h.annotations.ReviewAnnotationWithArtifacts(r.Context(), annotate.ReviewAnnotationActionInput{
		AnnotationID: r.PathValue("id"),
		ReviewState:  strings.TrimSpace(req.GetReviewState()),
		MailboxName:  strings.TrimSpace(req.GetMailboxName()),
		Comment:      protoCommentToAnnotate(req.GetComment()),
		GuidelineIDs: req.GetGuidelineIds(),
	})
	if err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, err.Error())
			return
		}
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, annotateAnnotationToProto(annotation))
}

func (h *appHandler) handleBatchReview(w http.ResponseWriter, r *http.Request) {
	req := &annotationuiv1.BatchReviewRequest{}
	if !decodeProtoJSONBody(w, r, req) {
		return
	}
	if len(req.GetIds()) == 0 {
		writeMessageError(w, http.StatusBadRequest, "ids must not be empty")
		return
	}
	if !isValidReviewState(req.GetReviewState()) {
		writeMessageError(w, http.StatusBadRequest, "reviewState must be one of to_review, reviewed, dismissed")
		return
	}
	if err := h.annotations.BatchReviewWithArtifacts(r.Context(), annotate.BatchReviewActionInput{
		IDs:          req.GetIds(),
		ReviewState:  strings.TrimSpace(req.GetReviewState()),
		AgentRunID:   strings.TrimSpace(req.GetAgentRunId()),
		MailboxName:  strings.TrimSpace(req.GetMailboxName()),
		Comment:      protoCommentToAnnotate(req.GetComment()),
		GuidelineIDs: req.GetGuidelineIds(),
	}); err != nil {
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

	req := &annotationuiv1.GroupListRequest{
		ReviewState: strings.TrimSpace(r.URL.Query().Get("reviewState")),
		SourceKind:  strings.TrimSpace(r.URL.Query().Get("sourceKind")),
		Limit:       int32(limit),
	}

	groups, err := h.annotations.ListGroups(r.Context(), annotate.ListGroupsFilter{
		ReviewState: req.GetReviewState(),
		SourceKind:  req.GetSourceKind(),
		Limit:       int(req.GetLimit()),
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, &annotationuiv1.GroupListResponse{Items: annotateGroupsToProto(groups)})
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
	writeProtoJSON(w, http.StatusOK, annotateGroupDetailToProto(annotate.GroupDetail{
		TargetGroup: *group,
		Members:     members,
	}))
}

func (h *appHandler) handleListLogs(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 200)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	req := &annotationuiv1.LogListRequest{
		SourceKind: strings.TrimSpace(r.URL.Query().Get("sourceKind")),
		AgentRunId: strings.TrimSpace(r.URL.Query().Get("agentRunId")),
		Limit:      int32(limit),
	}

	logs, err := h.annotations.ListLogs(r.Context(), annotate.ListLogsFilter{
		SourceKind: req.GetSourceKind(),
		AgentRunID: req.GetAgentRunId(),
		Limit:      int(req.GetLimit()),
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, &annotationuiv1.LogListResponse{Items: annotateLogsToProto(logs)})
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
	writeProtoJSON(w, http.StatusOK, annotateLogToProto(logEntry))
}

func (h *appHandler) handleListRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := h.annotations.ListRuns(r.Context())
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeProtoJSON(w, http.StatusOK, &annotationuiv1.AgentRunListResponse{Items: annotateRunsToProto(runs)})
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
	writeProtoJSON(w, http.StatusOK, annotateRunDetailToProto(run))
}

func isValidReviewState(reviewState string) bool {
	switch strings.TrimSpace(reviewState) {
	case annotate.ReviewStateToReview, annotate.ReviewStateReviewed, annotate.ReviewStateDismissed:
		return true
	default:
		return false
	}
}
