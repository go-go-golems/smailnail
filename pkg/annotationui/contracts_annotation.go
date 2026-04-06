package annotationui

import (
	"time"

	"github.com/go-go-golems/smailnail/pkg/annotate"
	annotationuiv1 "github.com/go-go-golems/smailnail/pkg/gen/smailnail/annotationui/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func formatProtoTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func databaseInfoToProto(info DatabaseInfo) *annotationuiv1.DatabaseInfo {
	return &annotationuiv1.DatabaseInfo{
		Driver: info.Driver,
		Target: info.Target,
		Mode:   info.Mode,
	}
}

func annotateAnnotationToProto(annotation *annotate.Annotation) *annotationuiv1.Annotation {
	if annotation == nil {
		return nil
	}
	return &annotationuiv1.Annotation{
		Id:           annotation.ID,
		TargetType:   annotation.TargetType,
		TargetId:     annotation.TargetID,
		Tag:          annotation.Tag,
		NoteMarkdown: annotation.NoteMarkdown,
		SourceKind:   annotation.SourceKind,
		SourceLabel:  annotation.SourceLabel,
		AgentRunId:   annotation.AgentRunID,
		ReviewState:  annotation.ReviewState,
		CreatedBy:    annotation.CreatedBy,
		CreatedAt:    formatProtoTime(annotation.CreatedAt),
		UpdatedAt:    formatProtoTime(annotation.UpdatedAt),
	}
}

func annotateAnnotationsToProto(annotations []annotate.Annotation) []*annotationuiv1.Annotation {
	ret := make([]*annotationuiv1.Annotation, 0, len(annotations))
	for i := range annotations {
		ret = append(ret, annotateAnnotationToProto(&annotations[i]))
	}
	return ret
}

func annotateGroupToProto(group *annotate.TargetGroup) *annotationuiv1.TargetGroup {
	if group == nil {
		return nil
	}
	return &annotationuiv1.TargetGroup{
		Id:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		SourceKind:  group.SourceKind,
		SourceLabel: group.SourceLabel,
		AgentRunId:  group.AgentRunID,
		ReviewState: group.ReviewState,
		CreatedBy:   group.CreatedBy,
		CreatedAt:   formatProtoTime(group.CreatedAt),
		UpdatedAt:   formatProtoTime(group.UpdatedAt),
	}
}

func annotateGroupsToProto(groups []annotate.TargetGroup) []*annotationuiv1.TargetGroup {
	ret := make([]*annotationuiv1.TargetGroup, 0, len(groups))
	for i := range groups {
		ret = append(ret, annotateGroupToProto(&groups[i]))
	}
	return ret
}

func annotateGroupMemberToProto(member *annotate.GroupMember) *annotationuiv1.GroupMember {
	if member == nil {
		return nil
	}
	return &annotationuiv1.GroupMember{
		GroupId:    member.GroupID,
		TargetType: member.TargetType,
		TargetId:   member.TargetID,
		AddedAt:    formatProtoTime(member.AddedAt),
	}
}

func annotateGroupMembersToProto(members []annotate.GroupMember) []*annotationuiv1.GroupMember {
	ret := make([]*annotationuiv1.GroupMember, 0, len(members))
	for i := range members {
		ret = append(ret, annotateGroupMemberToProto(&members[i]))
	}
	return ret
}

func annotateGroupDetailToProto(detail annotate.GroupDetail) *annotationuiv1.GroupDetail {
	return &annotationuiv1.GroupDetail{
		Id:          detail.ID,
		Name:        detail.Name,
		Description: detail.Description,
		SourceKind:  detail.SourceKind,
		SourceLabel: detail.SourceLabel,
		AgentRunId:  detail.AgentRunID,
		ReviewState: detail.ReviewState,
		CreatedBy:   detail.CreatedBy,
		CreatedAt:   formatProtoTime(detail.CreatedAt),
		UpdatedAt:   formatProtoTime(detail.UpdatedAt),
		Members:     annotateGroupMembersToProto(detail.Members),
	}
}

func annotateLogToProto(logEntry *annotate.AnnotationLog) *annotationuiv1.AnnotationLog {
	if logEntry == nil {
		return nil
	}
	return &annotationuiv1.AnnotationLog{
		Id:           logEntry.ID,
		LogKind:      logEntry.LogKind,
		Title:        logEntry.Title,
		BodyMarkdown: logEntry.BodyMarkdown,
		SourceKind:   logEntry.SourceKind,
		SourceLabel:  logEntry.SourceLabel,
		AgentRunId:   logEntry.AgentRunID,
		CreatedBy:    logEntry.CreatedBy,
		CreatedAt:    formatProtoTime(logEntry.CreatedAt),
	}
}

func annotateLogsToProto(logs []annotate.AnnotationLog) []*annotationuiv1.AnnotationLog {
	ret := make([]*annotationuiv1.AnnotationLog, 0, len(logs))
	for i := range logs {
		ret = append(ret, annotateLogToProto(&logs[i]))
	}
	return ret
}

func annotateRunSummaryToProto(run *annotate.AgentRunSummary) *annotationuiv1.AgentRunSummary {
	if run == nil {
		return nil
	}
	return &annotationuiv1.AgentRunSummary{
		RunId:           run.RunID,
		SourceLabel:     run.SourceLabel,
		SourceKind:      run.SourceKind,
		AnnotationCount: int32(run.AnnotationCount),
		PendingCount:    int32(run.PendingCount),
		ReviewedCount:   int32(run.ReviewedCount),
		DismissedCount:  int32(run.DismissedCount),
		LogCount:        int32(run.LogCount),
		GroupCount:      int32(run.GroupCount),
		StartedAt:       run.StartedAt,
		CompletedAt:     run.CompletedAt,
	}
}

func annotateRunsToProto(runs []annotate.AgentRunSummary) []*annotationuiv1.AgentRunSummary {
	ret := make([]*annotationuiv1.AgentRunSummary, 0, len(runs))
	for i := range runs {
		ret = append(ret, annotateRunSummaryToProto(&runs[i]))
	}
	return ret
}

func annotateRunDetailToProto(detail *annotate.AgentRunDetail) *annotationuiv1.AgentRunDetail {
	if detail == nil {
		return nil
	}
	return &annotationuiv1.AgentRunDetail{
		RunId:           detail.RunID,
		SourceLabel:     detail.SourceLabel,
		SourceKind:      detail.SourceKind,
		AnnotationCount: int32(detail.AnnotationCount),
		PendingCount:    int32(detail.PendingCount),
		ReviewedCount:   int32(detail.ReviewedCount),
		DismissedCount:  int32(detail.DismissedCount),
		LogCount:        int32(detail.LogCount),
		GroupCount:      int32(detail.GroupCount),
		StartedAt:       detail.StartedAt,
		CompletedAt:     detail.CompletedAt,
		Annotations:     annotateAnnotationsToProto(detail.Annotations),
		Logs:            annotateLogsToProto(detail.Logs),
		Groups:          annotateGroupsToProto(detail.Groups),
	}
}

func senderRowToProto(row SenderRow) *annotationuiv1.SenderRow {
	return &annotationuiv1.SenderRow{
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		Domain:          row.Domain,
		MessageCount:    int32(row.MessageCount),
		AnnotationCount: int32(row.AnnotationCount),
		Tags:            append([]string(nil), row.Tags...),
		HasUnsubscribe:  row.HasUnsubscribe,
	}
}

func senderRowsToProto(rows []SenderRow) []*annotationuiv1.SenderRow {
	ret := make([]*annotationuiv1.SenderRow, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, senderRowToProto(row))
	}
	return ret
}

func messagePreviewToProto(preview MessagePreview) *annotationuiv1.MessagePreview {
	return &annotationuiv1.MessagePreview{
		Uid:         preview.UID,
		Subject:     preview.Subject,
		Date:        preview.Date,
		SizeBytes:   int32(preview.SizeBytes),
		MailboxName: preview.MailboxName,
	}
}

func messagePreviewsToProto(previews []MessagePreview) []*annotationuiv1.MessagePreview {
	ret := make([]*annotationuiv1.MessagePreview, 0, len(previews))
	for _, preview := range previews {
		ret = append(ret, messagePreviewToProto(preview))
	}
	return ret
}

func senderDetailToProto(detail SenderDetail) *annotationuiv1.SenderDetail {
	return &annotationuiv1.SenderDetail{
		Email:           detail.Email,
		DisplayName:     detail.DisplayName,
		Domain:          detail.Domain,
		MessageCount:    int32(detail.MessageCount),
		AnnotationCount: int32(detail.AnnotationCount),
		Tags:            append([]string(nil), detail.Tags...),
		HasUnsubscribe:  detail.HasUnsubscribe,
		FirstSeen:       detail.FirstSeen,
		LastSeen:        detail.LastSeen,
		Annotations:     annotateAnnotationsToProto(detail.Annotations),
		Logs:            annotateLogsToProto(detail.Logs),
		RecentMessages:  messagePreviewsToProto(detail.RecentMessages),
	}
}

func savedQueryToProto(query SavedQuery) *annotationuiv1.SavedQuery {
	return &annotationuiv1.SavedQuery{
		Name:        query.Name,
		Folder:      query.Folder,
		Description: query.Description,
		Sql:         query.SQL,
	}
}

func savedQueriesToProto(queries []SavedQuery) []*annotationuiv1.SavedQuery {
	ret := make([]*annotationuiv1.SavedQuery, 0, len(queries))
	for _, query := range queries {
		ret = append(ret, savedQueryToProto(query))
	}
	return ret
}

func queryResultToProto(result QueryResult) (*annotationuiv1.QueryResult, error) {
	rows := make([]*structpb.Struct, 0, len(result.Rows))
	for _, row := range result.Rows {
		structured, err := structpb.NewStruct(row)
		if err != nil {
			return nil, err
		}
		rows = append(rows, structured)
	}
	return &annotationuiv1.QueryResult{
		Columns:    append([]string(nil), result.Columns...),
		Rows:       rows,
		DurationMs: result.DurationMs,
		RowCount:   int32(result.RowCount),
	}, nil
}
