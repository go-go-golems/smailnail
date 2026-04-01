package mirror

import "time"

const (
	DefaultSQLiteDBPath = "smailnail-mirror.sqlite"
	DefaultMirrorRoot   = "smailnail-mirror"
)

const (
	SearchModeFTS5 = "fts5"
)

type DatabaseInfo struct {
	Driver string `json:"driver"`
	Path   string `json:"path"`
}

type MailboxSyncState struct {
	AccountKey  string     `db:"account_key" json:"accountKey"`
	MailboxName string     `db:"mailbox_name" json:"mailboxName"`
	UIDValidity uint32     `db:"uidvalidity" json:"uidValidity"`
	HighestUID  uint32     `db:"highest_uid" json:"highestUid"`
	LastUIDNext uint32     `db:"last_uidnext" json:"lastUidNext"`
	LastSyncAt  *time.Time `db:"last_sync_at" json:"lastSyncAt,omitempty"`
	Status      string     `db:"status" json:"status"`
}

type MessageRecord struct {
	ID             int64      `db:"id" json:"id"`
	AccountKey     string     `db:"account_key" json:"accountKey"`
	MailboxName    string     `db:"mailbox_name" json:"mailboxName"`
	UIDValidity    uint32     `db:"uidvalidity" json:"uidValidity"`
	UID            uint32     `db:"uid" json:"uid"`
	MessageID      string     `db:"message_id" json:"messageId"`
	InternalDate   string     `db:"internal_date" json:"internalDate"`
	SentDate       string     `db:"sent_date" json:"sentDate"`
	Subject        string     `db:"subject" json:"subject"`
	FromSummary    string     `db:"from_summary" json:"fromSummary"`
	ToSummary      string     `db:"to_summary" json:"toSummary"`
	CCSummary      string     `db:"cc_summary" json:"ccSummary"`
	SizeBytes      int64      `db:"size_bytes" json:"sizeBytes"`
	FlagsJSON      string     `db:"flags_json" json:"flagsJson"`
	HeadersJSON    string     `db:"headers_json" json:"headersJson"`
	PartsJSON      string     `db:"parts_json" json:"partsJson"`
	BodyText       string     `db:"body_text" json:"bodyText"`
	BodyHTML       string     `db:"body_html" json:"bodyHTML"`
	SearchText     string     `db:"search_text" json:"searchText"`
	RawPath        string     `db:"raw_path" json:"rawPath"`
	RawSHA256      string     `db:"raw_sha256" json:"rawSHA256"`
	HasAttachments bool       `db:"has_attachments" json:"hasAttachments"`
	RemoteDeleted  bool       `db:"remote_deleted" json:"remoteDeleted"`
	FirstSeenAt    *time.Time `db:"first_seen_at" json:"firstSeenAt,omitempty"`
	LastSyncedAt   *time.Time `db:"last_synced_at" json:"lastSyncedAt,omitempty"`
}

type BootstrapReport struct {
	Database        DatabaseInfo `json:"database"`
	MirrorRoot      string       `json:"mirrorRoot"`
	SearchMode      string       `json:"searchMode"`
	FTSAvailable    bool         `json:"ftsAvailable"`
	FTSStatus       string       `json:"ftsStatus"`
	SchemaVersion   int          `json:"schemaVersion"`
	PrintPlan       bool         `json:"printPlan"`
	SelectedMailbox string       `json:"selectedMailbox"`
	AllMailboxes    bool         `json:"allMailboxes"`
	BatchSize       int          `json:"batchSize"`
	ResetState      bool         `json:"resetState"`
	ReconcileFull   bool         `json:"reconcileFull"`
}

type RawMessageResult struct {
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	BytesWritten int    `json:"bytesWritten"`
	Reused       bool   `json:"reused"`
}

type MailboxSyncResult struct {
	MailboxName        string `json:"mailboxName"`
	UIDValidity        uint32 `json:"uidValidity"`
	UIDNext            uint32 `json:"uidNext"`
	PreviousHighUID    uint32 `json:"previousHighUid"`
	HighestUID         uint32 `json:"highestUid"`
	FetchedMessages    int    `json:"fetchedMessages"`
	StoredMessages     int    `json:"storedMessages"`
	RawFilesWritten    int    `json:"rawFilesWritten"`
	ReusedFileWrites   int    `json:"reusedFileWrites"`
	TombstonedMessages int    `json:"tombstonedMessages"`
	RestoredMessages   int    `json:"restoredMessages"`
	ReconcileApplied   bool   `json:"reconcileApplied"`
	ResetApplied       bool   `json:"resetApplied"`
}

type SyncReport struct {
	AccountKey         string              `json:"accountKey"`
	MailboxesPlanned   int                 `json:"mailboxesPlanned"`
	MailboxesSynced    int                 `json:"mailboxesSynced"`
	MessagesFetched    int                 `json:"messagesFetched"`
	MessagesStored     int                 `json:"messagesStored"`
	RawFilesWritten    int                 `json:"rawFilesWritten"`
	ReusedFileWrites   int                 `json:"reusedFileWrites"`
	TombstonedMessages int                 `json:"tombstonedMessages"`
	RestoredMessages   int                 `json:"restoredMessages"`
	Mailboxes          []MailboxSyncResult `json:"mailboxes"`
}
