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
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
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
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   2,
		StopOnError: true,
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
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   2,
		StopOnError: true,
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

func TestServiceSyncHonorsMaxMessages(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 4}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Alpha"),
		2: newFetchedMessage(2, "Beta"),
		3: newFetchedMessage(3, "Gamma"),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}
	service.now = func() time.Time { return time.Date(2026, 4, 1, 19, 50, 0, 0, time.UTC) }

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   10,
		MaxMessages: 2,
		StopOnError: true,
	})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if report.MessagesFetched != 2 || report.MessagesStored != 2 {
		t.Fatalf("expected max-limited sync to fetch/store 2 messages, got %+v", report)
	}
	if !report.MaxMessagesReached {
		t.Fatalf("expected sync report to record max message limit, got %+v", report)
	}
	if got := countMessages(t, store.db, "INBOX"); got != 2 {
		t.Fatalf("expected 2 mirrored messages after limited sync, got %d", got)
	}

	state, err := loadMailboxSyncState(t.Context(), store.db, AccountKey("localhost", 993, "a"), "INBOX")
	if err != nil {
		t.Fatalf("loadMailboxSyncState() error = %v", err)
	}
	if state == nil || state.HighestUID != 2 {
		t.Fatalf("expected max-limited sync to checkpoint at UID 2, got %+v", state)
	}
}

func TestServiceSyncHonorsSinceDays(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	baseNow := time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC)
	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 4}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessageAt(1, "Recent", baseNow.Add(-24*time.Hour)),
		2: newFetchedMessageAt(2, "Still Recent", baseNow.Add(-48*time.Hour)),
		3: newFetchedMessageAt(3, "Old", baseNow.Add(-10*24*time.Hour)),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}
	service.now = func() time.Time { return baseNow }

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   10,
		SinceDays:   3,
		MaxMessages: 0,
		StopOnError: true,
	})
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if report.MessagesFetched != 2 || report.MessagesStored != 2 {
		t.Fatalf("expected recent-only sync to fetch/store 2 messages, got %+v", report)
	}
	if got := countMessages(t, store.db, "INBOX"); got != 2 {
		t.Fatalf("expected 2 mirrored messages after since-days sync, got %d", got)
	}
	if hasMessage(t, store.db, "INBOX", 3) {
		t.Fatalf("expected old message UID 3 to be excluded by since-days")
	}
}

func TestServiceSyncResetsOnUIDValidityChange(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
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
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   10,
		StopOnError: true,
	}); err != nil {
		t.Fatalf("initial Sync() error = %v", err)
	}

	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 77, UIDNext: 2}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Replacement"),
	}

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   10,
		StopOnError: true,
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
	criteria, ok := newUIDSearchCriteria(2, 3, nil)
	if ok {
		t.Fatalf("expected no search when highest UID already reaches UIDNEXT, got %+v", criteria)
	}

	criteria, ok = newUIDSearchCriteria(2, 5, nil)
	if !ok || criteria == nil || criteria.UID == nil {
		t.Fatalf("expected bounded UID search criteria, got %+v", criteria)
	}
	if got := criteria.UID.String(); got != "3:4" {
		t.Fatalf("expected UID range 3:4, got %q", got)
	}
}

func TestSinceDaysCutoff(t *testing.T) {
	baseNow := time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC)
	cutoff := sinceDaysCutoff(baseNow, 3)
	if cutoff == nil {
		t.Fatalf("expected cutoff for positive since-days")
	}
	if got := cutoff.Format(time.RFC3339); got != "2026-03-29T20:00:00Z" {
		t.Fatalf("unexpected cutoff %q", got)
	}
	if cutoff := sinceDaysCutoff(baseNow, 0); cutoff != nil {
		t.Fatalf("expected nil cutoff when since-days is zero")
	}
}

func TestResolveMailboxesAppliesIncludeAndExcludePatterns(t *testing.T) {
	service := NewService(openTestStore(t))
	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{
		{Name: "INBOX"},
		{Name: "Archive/2026"},
		{Name: "Archive/2025"},
		{Name: "Spam"},
		{Name: "Archive/Hidden", Flags: []string{"\\Noselect"}},
	}

	mailboxes, err := service.resolveMailboxes(session, SyncOptions{
		AllMailboxes:          true,
		MailboxPattern:        "Archive/*",
		ExcludeMailboxPattern: "*/2025",
	})
	if err != nil {
		t.Fatalf("resolveMailboxes() error = %v", err)
	}
	expected := []string{"Archive/2026"}
	if fmt.Sprintf("%v", mailboxes) != fmt.Sprintf("%v", expected) {
		t.Fatalf("expected filtered mailboxes %v, got %v", expected, mailboxes)
	}
}

func TestServiceSyncStopOnErrorFailsFast(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "Broken"}, {Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 2}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Alpha"),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}

	_, err := service.Sync(t.Context(), SyncOptions{
		Server:       "localhost",
		Port:         993,
		Username:     "a",
		Password:     "pass",
		Insecure:     true,
		AllMailboxes: true,
		MirrorRoot:   root,
		BatchSize:    10,
		StopOnError:  true,
	})
	if err == nil {
		t.Fatalf("expected fail-fast sync to return an error")
	}
}

func TestServiceSyncContinuesWhenStopOnErrorDisabled(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "Broken"}, {Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 2}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Alpha"),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:       "localhost",
		Port:         993,
		Username:     "a",
		Password:     "pass",
		Insecure:     true,
		AllMailboxes: true,
		MirrorRoot:   root,
		BatchSize:    10,
		StopOnError:  false,
	})
	if err != nil {
		t.Fatalf("expected continue-on-error sync to succeed, got %v", err)
	}
	if report.MailboxErrors != 1 {
		t.Fatalf("expected exactly one mailbox error, got %+v", report)
	}
	if report.MailboxesSynced != 1 {
		t.Fatalf("expected one successful mailbox sync, got %+v", report)
	}
	if got := countMessages(t, store.db, "INBOX"); got != 1 {
		t.Fatalf("expected INBOX to still sync after Broken failed, got %d mirrored messages", got)
	}
	if fmt.Sprintf("%v", report.FailedMailboxes) != "[Broken]" {
		t.Fatalf("expected failed mailbox list to contain Broken, got %+v", report.FailedMailboxes)
	}
}

func TestServiceSyncReconcileTombstonesMissingMessages(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	session := newFakeIMAPSession()
	session.mailboxes = []mailruntime.MailboxInfo{{Name: "INBOX"}}
	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 4}
	session.messages["INBOX"] = map[uint32]*mailruntime.FetchedMessage{
		1: newFetchedMessage(1, "Alpha"),
		2: newFetchedMessage(2, "Beta"),
		3: newFetchedMessage(3, "Gamma"),
	}

	service := NewService(store)
	service.dial = func(_ context.Context, _ mailruntime.IMAPOptions) (imapSession, error) {
		return session, nil
	}
	service.now = func() time.Time { return time.Date(2026, 4, 1, 20, 30, 0, 0, time.UTC) }

	if _, err := service.Sync(t.Context(), SyncOptions{
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   10,
		StopOnError: true,
	}); err != nil {
		t.Fatalf("initial Sync() error = %v", err)
	}

	session.statuses["INBOX"] = &mailruntime.MailboxStatus{UIDValidity: 21, UIDNext: 4}
	delete(session.messages["INBOX"], 2)

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:        "localhost",
		Port:          993,
		Username:      "a",
		Password:      "pass",
		Insecure:      true,
		Mailbox:       "INBOX",
		MirrorRoot:    root,
		BatchSize:     10,
		ReconcileFull: true,
		StopOnError:   true,
	})
	if err != nil {
		t.Fatalf("reconcile Sync() error = %v", err)
	}

	if len(report.Mailboxes) != 1 {
		t.Fatalf("expected 1 mailbox report, got %+v", report.Mailboxes)
	}
	if !report.Mailboxes[0].ReconcileApplied {
		t.Fatalf("expected reconcile to be recorded, got %+v", report.Mailboxes[0])
	}
	if report.Mailboxes[0].TombstonedMessages != 1 || report.TombstonedMessages != 1 {
		t.Fatalf("expected one tombstoned message, got mailbox=%+v report=%+v", report.Mailboxes[0], report)
	}
	if got := messageRemoteDeleted(t, store.db, "INBOX", 2); !got {
		t.Fatalf("expected UID 2 to be marked remote_deleted")
	}
	if got := messageRemoteDeleted(t, store.db, "INBOX", 1); got {
		t.Fatalf("expected UID 1 to remain active")
	}
}

func TestServiceSyncReconcileRestoresPresentMessages(t *testing.T) {
	store := openTestStore(t)
	root := t.TempDir()
	if _, err := store.Bootstrap(t.Context(), root); err != nil {
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
	service.now = func() time.Time { return time.Date(2026, 4, 1, 20, 45, 0, 0, time.UTC) }

	if _, err := service.Sync(t.Context(), SyncOptions{
		Server:      "localhost",
		Port:        993,
		Username:    "a",
		Password:    "pass",
		Insecure:    true,
		Mailbox:     "INBOX",
		MirrorRoot:  root,
		BatchSize:   10,
		StopOnError: true,
	}); err != nil {
		t.Fatalf("initial Sync() error = %v", err)
	}

	if _, err := store.db.Exec(`UPDATE messages SET remote_deleted = TRUE WHERE mailbox_name = ? AND uid = ?`, "INBOX", 2); err != nil {
		t.Fatalf("seed remote_deleted state error = %v", err)
	}

	report, err := service.Sync(t.Context(), SyncOptions{
		Server:        "localhost",
		Port:          993,
		Username:      "a",
		Password:      "pass",
		Insecure:      true,
		Mailbox:       "INBOX",
		MirrorRoot:    root,
		BatchSize:     10,
		ReconcileFull: true,
		StopOnError:   true,
	})
	if err != nil {
		t.Fatalf("reconcile Sync() error = %v", err)
	}

	if report.Mailboxes[0].RestoredMessages != 1 || report.RestoredMessages != 1 {
		t.Fatalf("expected one restored message, got mailbox=%+v report=%+v", report.Mailboxes[0], report)
	}
	if got := messageRemoteDeleted(t, store.db, "INBOX", 2); got {
		t.Fatalf("expected UID 2 to be restored")
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

func hasMessage(t *testing.T, db *sqlx.DB, mailboxName string, uid uint32) bool {
	t.Helper()

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM messages WHERE mailbox_name = ? AND uid = ?`, mailboxName, uid); err != nil {
		t.Fatalf("count message error = %v", err)
	}
	return count > 0
}

func messageRemoteDeleted(t *testing.T, db *sqlx.DB, mailboxName string, uid uint32) bool {
	t.Helper()

	var remoteDeleted bool
	if err := db.Get(&remoteDeleted, `SELECT remote_deleted FROM messages WHERE mailbox_name = ? AND uid = ?`, mailboxName, uid); err != nil {
		t.Fatalf("lookup remote_deleted error = %v", err)
	}
	return remoteDeleted
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
		msg := msgs[uid]
		imapUID := imap.UID(uid)
		if criteria != nil && criteria.UID != nil && !criteria.UID.Contains(imapUID) {
			continue
		}
		if criteria != nil && criteria.Since != nil {
			msgTime, err := time.Parse(time.RFC3339, msg.InternalDate)
			if err != nil {
				return nil, err
			}
			if msgTime.Before(*criteria.Since) {
				continue
			}
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
	return newFetchedMessageAt(uid, subject, time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC))
}

func newFetchedMessageAt(uid uint32, subject string, msgTime time.Time) *mailruntime.FetchedMessage {
	raw := []byte("From: Tester <test@example.com>\r\nTo: User <user@example.com>\r\nSubject: " + subject + "\r\n\r\nBody for " + subject + "\r\n")
	return &mailruntime.FetchedMessage{
		UID:          uid,
		Flags:        []string{"\\Seen"},
		Size:         int64(len(raw)),
		InternalDate: msgTime.Format(time.RFC3339),
		Envelope: &mailruntime.MessageEnvelope{
			Date:      msgTime.Format(time.RFC3339),
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
