package mirror

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/jmoiron/sqlx"

	"github.com/go-go-golems/smailnail/pkg/mailruntime"
)

func TestServiceSyncPersistsIncrementalMessages(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root, SearchModeBasic); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 3}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Alpha"),
		2: newFetchedMessage(2, "Beta"),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}
	fixedNow := time.Date(2026, 4, 1, 19, 45, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:     "localhost",
		Port:       993,
		Username:   "a",
		Password:   "pass",
		Insecure:   true,
		Mailbox:    "INBOX",
		MirrorRoot: root,
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if report.MessagesFetched != 2 || report.MessagesStored != 2 {
		t.Fatalf("unexpected report counts: %+v", report)
	}

	state, err := loadMailboxSyncState(t.Context(), store.db, AccountKey("localhost", 993, "a"), "INBOX")
	if err != nil {
		t.Fatalf("loadMailboxSyncState() error = %v", err)
	}
	if state == nil || state.HighestUID != 2 {
		t.Fatalf("unexpected mailbox state: %+v", state)
	}

	if got := countMessages(t, store.db, "INBOX"); got != 2 {
		t.Fatalf("expected 2 mirrored messages, got %d", got)
	}

	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 4}
	session.messages["INBOX"][3] = newFetchedMessage(3, "Gamma")

	report, err = service.Sync(t.Context(), SyncOptions{
		Server:     "localhost",
		Port:       993,
		Username:   "a",
		Password:   "pass",
		Insecure:   true,
		Mailbox:    "INBOX",
		MirrorRoot: root,
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("Sync() second error = %v", err)
	}
	if report.MessagesFetched != 1 || report.MessagesStored != 1 {
		t.Fatalf("expected incremental sync to fetch 1 message, got %+v", report)
	}
	if got := countMessages(t, store.db, "INBOX"); got != 3 {
		t.Fatalf("expected 3 mirrored messages after incremental sync, got %d", got)
	}

	rawPath := filepath.Join(root, RawMessagePath(AccountKey("localhost", 993, "a"), "INBOX", 21, 3))
	assertFileExists(t, rawPath)
}

func TestServiceSyncResetsOnUIDValidityChange(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root, SearchModeBasic); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 10, UIDNext: 2}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Original"),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}
	service.now = func() time.Time { return time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC) }

	if _, err := service.Sync(t.Context(), SyncOptions{
		Server:     "localhost",
		Port:       993,
		Username:   "a",
		Password:   "pass",
		Insecure:   true,
		Mailbox:    "INBOX",
		MirrorRoot: root,
		BatchSize:  10,
	}); err != nil {
		t.Fatalf("initial Sync() error = %v", err)
	}

	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 77, UIDNext: 2}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Replacement"),
	}

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:     "localhost",
		Port:       993,
		Username:   "a",
		Password:   "pass",
		Insecure:   true,
		Mailbox:    "INBOX",
		MirrorRoot: root,
		BatchSize:  10,
	})
	if err != nil {
		t.Fatalf("second Sync() error = %v", err)
	}
	if len(report.Mailboxes) != 1 || !report.Mailboxes[0].ResetApplied {
		t.Fatalf("expected UIDVALIDITY reset to be recorded, got %+v", report.Mailboxes)
	}

	if got := countMessages(t, store.db, "INBOX"); got != 1 {
		t.Fatalf("expected reset mailbox to contain exactly one message, got %d", got)
	}

	state, err := loadMailboxSyncState(t.Context(), store.db, AccountKey("localhost", 993, "a"), "INBOX")
	if err != nil {
		t.Fatalf("loadMailboxSyncState() error = %v", err)
	}
	if state == nil || state.UIDValidity != 77 {
		t.Fatalf("expected updated UIDVALIDITY state, got %+v", state)
	}
}

func TestNewUIDSearchCriteriaUsesUIDNextBoundary(t *testing.T) {
	criteria, ok := newUIDSearchCriteria(2, 3)
	if ok {
		t.Fatalf("expected no search when highest UID already reaches UIDNEXT, got %+v", criteria)
	}

	criteria, ok = newUIDSearchCriteria(2, 5)
	if !ok || criteria == nil || criteria.UID == nil {
		t.Fatalf("expected bounded UID search criteria, got %+v", criteria)
	}
	if got := criteria.UID.String(); got != "3:4" {
		t.Fatalf("expected UID range 3:4, got %q", got)
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()

	db := sqlx.MustOpen("sqlite3", t.TempDir()+"/mirror.sqlite")
	t.Cleanup(func() {
		_ = db.Close()
	})
	return &Store{
		db:   db,
		path: "test.sqlite",
	}
}

func countMessages(t *testing.T, db *sqlx.DB, mailboxName string) int {
	t.Helper()

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM messages WHERE mailbox_name = ?`, mailboxName); err != nil {
		t.Fatalf("count messages error = %v", err)
	}
	return count
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
}

type fakeIMAPSession struct {
	mailboxes []mailruntime.MailboxInfo
	statuses  map[string]*mailruntime.MailboxStatus
	messages  map[string]map[uint32]*mailruntime.FetchedMessage
	selected  string
}

func newFakeIMAPSession() *fakeIMAPSession {
	return &fakeIMAPSession{
		statuses: make(map[string]*mailruntime.MailboxStatus),
		messages: make(map[string]map[uint32]*mailruntime.FetchedMessage),
	}
}

func (f *fakeIMAPSession) List(_ string) ([]mailruntime.MailboxInfo, error) {
	return append([]mailruntime.MailboxInfo(nil), f.mailboxes...), nil
}

func (f *fakeIMAPSession) Status(name string) (*mailruntime.MailboxStatus, error) {
	status, ok := f.statuses[name]
	if !ok {
		return nil, fmt.Errorf("unknown mailbox %s", name)
	}
	statusCopy := *status
	return &statusCopy, nil
}

func (f *fakeIMAPSession) SelectMailbox(name string, _ bool) (*imap.SelectData, error) {
	if _, ok := f.messages[name]; !ok {
		return nil, fmt.Errorf("unknown mailbox %s", name)
	}
	f.selected = name
	return &imap.SelectData{}, nil
}

func (f *fakeIMAPSession) UnselectMailbox() error {
	f.selected = ""
	return nil
}

func (f *fakeIMAPSession) Search(criteria *mailruntime.SearchCriteria) ([]imap.UID, error) {
	msgs := f.messages[f.selected]
	ret := make([]imap.UID, 0, len(msgs))
	for uid := range msgs {
		imapUID := imap.UID(uid)
		if criteria != nil && criteria.UID != nil && !criteria.UID.Contains(imapUID) {
			continue
		}
		ret = append(ret, imapUID)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret, nil
}

func (f *fakeIMAPSession) Fetch(uids []imap.UID, _ []mailruntime.FetchField) ([]*mailruntime.FetchedMessage, error) {
	msgs := f.messages[f.selected]
	ret := make([]*mailruntime.FetchedMessage, 0, len(uids))
	for _, uid := range uids {
		msg, ok := msgs[uint32(uid)]
		if !ok {
			return nil, fmt.Errorf("unknown uid %d", uid)
		}
		msgCopy := *msg
		ret = append(ret, &msgCopy)
	}
	return ret, nil
}

func (f *fakeIMAPSession) Logout() error {
	return nil
}

func newFetchedMessage(uid uint32, subject string) *mailruntime.FetchedMessage {
	raw := []byte("From: Tester <test@example.com>\r\nTo: User <user@example.com>\r\nSubject: " + subject + "\r\n\r\nBody for " + subject + "\r\n")
	return &mailruntime.FetchedMessage{
		UID:          uid,
		Flags:        []string{"\\Seen"},
		Size:         int64(len(raw)),
		InternalDate: time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Envelope: &mailruntime.MessageEnvelope{
			Date:      time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Subject:   subject,
			From:      []string{"Tester <test@example.com>"},
			To:        []string{"User <user@example.com>"},
			MessageID: fmt.Sprintf("<msg-%d@example.com>", uid),
		},
		Headers: map[string]string{
			"Subject": subject,
		},
		BodyText:    "Body for " + subject,
		BodyRaw:     raw,
		Attachments: nil,
	}
}
