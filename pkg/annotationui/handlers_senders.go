package annotationui

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type senderRowRecord struct {
	Email           string `db:"email"`
	DisplayName     string `db:"display_name"`
	Domain          string `db:"domain"`
	MessageCount    int    `db:"message_count"`
	AnnotationCount int    `db:"annotation_count"`
	TagsCSV         string `db:"tags_csv"`
	HasUnsubscribe  bool   `db:"has_unsubscribe"`
}

type senderDetailRecord struct {
	Email          string `db:"email"`
	DisplayName    string `db:"display_name"`
	Domain         string `db:"domain"`
	MessageCount   int    `db:"message_count"`
	FirstSeen      string `db:"first_seen"`
	LastSeen       string `db:"last_seen"`
	HasUnsubscribe bool   `db:"has_unsubscribe"`
}

func (h *appHandler) handleListSenders(w http.ResponseWriter, r *http.Request) {
	limit, err := parseLimitQuery(r, "limit", 200)
	if err != nil {
		writeMessageError(w, http.StatusBadRequest, err.Error())
		return
	}

	hasAnnotations := false
	if raw := strings.TrimSpace(r.URL.Query().Get("hasAnnotations")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			writeMessageError(w, http.StatusBadRequest, "hasAnnotations must be a boolean")
			return
		}
		hasAnnotations = parsed
	}

	query := `
SELECT
	s.email,
	s.display_name,
	s.domain,
	s.msg_count AS message_count,
	COUNT(DISTINCT a.id) AS annotation_count,
	COALESCE(GROUP_CONCAT(DISTINCT NULLIF(a.tag, '')), '') AS tags_csv,
	s.has_list_unsubscribe AS has_unsubscribe
FROM senders s
LEFT JOIN annotations a ON a.target_type = 'sender' AND a.target_id = s.email
WHERE 1 = 1`

	args := make([]any, 0, 4)
	if domain := strings.TrimSpace(r.URL.Query().Get("domain")); domain != "" {
		query += ` AND s.domain = ?`
		args = append(args, domain)
	}
	if tag := strings.TrimSpace(r.URL.Query().Get("tag")); tag != "" {
		query += ` AND a.tag = ?`
		args = append(args, tag)
	}
	query += `
GROUP BY s.email, s.display_name, s.domain, s.msg_count, s.has_list_unsubscribe`
	if hasAnnotations {
		query += ` HAVING COUNT(DISTINCT a.id) > 0`
	}
	query += ` ORDER BY s.msg_count DESC, s.email ASC LIMIT ?`
	args = append(args, limit)

	rows := []senderRowRecord{}
	if err := h.db.SelectContext(r.Context(), &rows, h.db.Rebind(query), args...); err != nil {
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "list senders").Error())
		return
	}

	ret := make([]SenderRow, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, SenderRow{
			Email:           row.Email,
			DisplayName:     row.DisplayName,
			Domain:          row.Domain,
			MessageCount:    row.MessageCount,
			AnnotationCount: row.AnnotationCount,
			Tags:            splitTags(row.TagsCSV),
			HasUnsubscribe:  row.HasUnsubscribe,
		})
	}

	writeJSON(w, http.StatusOK, ret)
}

func (h *appHandler) handleGetSender(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.PathValue("email"))
	if email == "" {
		writeNotFound(w, "sender not found")
		return
	}

	record := senderDetailRecord{}
	query := h.db.Rebind(`
SELECT
	email,
	display_name,
	domain,
	msg_count AS message_count,
	first_seen_date AS first_seen,
	last_seen_date AS last_seen,
	has_list_unsubscribe AS has_unsubscribe
FROM senders
WHERE email = ?`)
	if err := h.db.GetContext(r.Context(), &record, query, email); err != nil {
		if isNotFoundError(err) {
			writeNotFound(w, "sender not found")
			return
		}
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "get sender").Error())
		return
	}

	annotations, err := h.annotations.ListAnnotations(r.Context(), annotate.ListAnnotationsFilter{
		TargetType: "sender",
		TargetID:   email,
		Limit:      500,
	})
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logs, err := h.listSenderLogs(r.Context(), annotations)
	if err != nil {
		writeMessageError(w, http.StatusInternalServerError, err.Error())
		return
	}

	recentMessages := []MessagePreview{}
	messageQuery := h.db.Rebind(`
SELECT
	uid,
	subject,
	COALESCE(NULLIF(sent_date, ''), NULLIF(internal_date, '')) AS date,
	size_bytes
FROM messages
WHERE sender_email = ?
ORDER BY COALESCE(NULLIF(sent_date, ''), NULLIF(internal_date, '')) DESC, uid DESC
LIMIT 20`)
	if err := h.db.SelectContext(r.Context(), &recentMessages, messageQuery, email); err != nil {
		writeMessageError(w, http.StatusInternalServerError, errors.Wrap(err, "list sender messages").Error())
		return
	}

	tags := make([]string, 0)
	seenTags := map[string]struct{}{}
	for _, annotation := range annotations {
		tag := strings.TrimSpace(annotation.Tag)
		if tag == "" {
			continue
		}
		if _, ok := seenTags[tag]; ok {
			continue
		}
		seenTags[tag] = struct{}{}
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	writeJSON(w, http.StatusOK, SenderDetail{
		SenderRow: SenderRow{
			Email:           record.Email,
			DisplayName:     record.DisplayName,
			Domain:          record.Domain,
			MessageCount:    record.MessageCount,
			AnnotationCount: len(annotations),
			Tags:            tags,
			HasUnsubscribe:  record.HasUnsubscribe,
		},
		FirstSeen:      record.FirstSeen,
		LastSeen:       record.LastSeen,
		Annotations:    annotations,
		Logs:           logs,
		RecentMessages: recentMessages,
	})
}

func (h *appHandler) listSenderLogs(ctx context.Context, annotations []annotate.Annotation) ([]annotate.AnnotationLog, error) {
	runIDs := make([]string, 0)
	seen := map[string]struct{}{}
	for _, annotation := range annotations {
		runID := strings.TrimSpace(annotation.AgentRunID)
		if runID == "" {
			continue
		}
		if _, ok := seen[runID]; ok {
			continue
		}
		seen[runID] = struct{}{}
		runIDs = append(runIDs, runID)
	}
	if len(runIDs) == 0 {
		return []annotate.AnnotationLog{}, nil
	}

	query, args, err := sqlx.In(`SELECT * FROM annotation_logs
		WHERE agent_run_id IN (?)
		ORDER BY created_at DESC, id DESC`, runIDs)
	if err != nil {
		return nil, errors.Wrap(err, "build sender logs query")
	}
	ret := []annotate.AnnotationLog{}
	if err := h.db.SelectContext(ctx, &ret, h.db.Rebind(query), args...); err != nil {
		return nil, errors.Wrap(err, "list sender logs")
	}
	return ret, nil
}

func splitTags(tagsCSV string) []string {
	if strings.TrimSpace(tagsCSV) == "" {
		return []string{}
	}
	parts := strings.Split(tagsCSV, ",")
	ret := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		ret = append(ret, part)
	}
	sort.Strings(ret)
	return ret
}
