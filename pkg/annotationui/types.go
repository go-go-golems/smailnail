package annotationui

import "github.com/go-go-golems/smailnail/pkg/annotate"

type DatabaseInfo struct {
	Driver string `json:"driver"`
	Target string `json:"target"`
	Mode   string `json:"mode"`
}

type SenderRow struct {
	Email           string   `json:"email"`
	DisplayName     string   `json:"displayName"`
	Domain          string   `json:"domain"`
	MessageCount    int      `json:"messageCount"`
	AnnotationCount int      `json:"annotationCount"`
	Tags            []string `json:"tags"`
	HasUnsubscribe  bool     `json:"hasUnsubscribe"`
}

type SenderDetail struct {
	SenderRow
	FirstSeen      string                   `json:"firstSeen"`
	LastSeen       string                   `json:"lastSeen"`
	Annotations    []annotate.Annotation    `json:"annotations"`
	Logs           []annotate.AnnotationLog `json:"logs"`
	RecentMessages []MessagePreview         `json:"recentMessages"`
}

type MessagePreview struct {
	UID       uint32 `db:"uid" json:"uid"`
	Subject   string `db:"subject" json:"subject"`
	Date      string `db:"date" json:"date"`
	SizeBytes int    `db:"size_bytes" json:"sizeBytes"`
}

type SavedQuery struct {
	Name        string `json:"name"`
	Folder      string `json:"folder"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
}

type QueryResult struct {
	Columns    []string         `json:"columns"`
	Rows       []map[string]any `json:"rows"`
	DurationMs int64            `json:"durationMs"`
	RowCount   int              `json:"rowCount"`
}
