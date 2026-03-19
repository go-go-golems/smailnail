package smailnailjs

import (
	"context"
	"testing"
	"time"

	"github.com/go-go-golems/smailnail/pkg/dsl"
)

type fakeSession struct {
	mailbox string
	closed  bool
}

func (s *fakeSession) Mailbox() string {
	return s.mailbox
}

func (s *fakeSession) Capabilities() map[string]bool {
	return map[string]bool{"uidplus": true}
}

func (s *fakeSession) List(pattern string) ([]MailboxInfo, error) {
	return []MailboxInfo{{Name: "INBOX"}}, nil
}

func (s *fakeSession) Status(name string) (*MailboxStatus, error) {
	return &MailboxStatus{Messages: 3, UIDNext: 4}, nil
}

func (s *fakeSession) SelectMailbox(name string, readOnly bool) (*MailboxSelection, error) {
	s.mailbox = name
	return &MailboxSelection{Name: name, ReadOnly: readOnly}, nil
}

func (s *fakeSession) Search(criteria *SearchCriteria) ([]uint32, error) {
	return []uint32{101, 202}, nil
}

func (s *fakeSession) Fetch(uids []uint32, fields []FetchField) ([]*FetchedMessage, error) {
	return []*FetchedMessage{{UID: 101}}, nil
}

func (s *fakeSession) AddFlags(uids []uint32, flags []string, silent bool) error {
	return nil
}

func (s *fakeSession) RemoveFlags(uids []uint32, flags []string, silent bool) error {
	return nil
}

func (s *fakeSession) SetFlags(uids []uint32, flags []string, silent bool) error {
	return nil
}

func (s *fakeSession) Move(uids []uint32, dest string) error {
	return nil
}

func (s *fakeSession) Copy(uids []uint32, dest string) error {
	return nil
}

func (s *fakeSession) Delete(uids []uint32, expunge bool) error {
	return nil
}

func (s *fakeSession) Expunge(uids []uint32) error {
	return nil
}

func (s *fakeSession) Append(mailbox string, message []byte, flags []string, date *time.Time) (uint32, error) {
	return 9001, nil
}

func (s *fakeSession) Close() {
	s.closed = true
}

type fakeDialer struct {
	gotOpts ConnectOptions
	session *fakeSession
}

func (d *fakeDialer) Dial(_ context.Context, opts ConnectOptions) (Session, error) {
	d.gotOpts = normalizeConnectOptions(opts)
	if d.session == nil {
		d.session = &fakeSession{mailbox: d.gotOpts.Mailbox}
	}
	return d.session, nil
}

type fakeSieveSession struct {
	closed bool
}

func (s *fakeSieveSession) Capabilities() SieveCapabilities {
	return SieveCapabilities{Implementation: "Fake"}
}

func (s *fakeSieveSession) ListScripts() ([]ScriptInfo, error) {
	return []ScriptInfo{{Name: "active", Active: true}}, nil
}

func (s *fakeSieveSession) GetScript(name string) (string, error) {
	return "keep;", nil
}

func (s *fakeSieveSession) PutScript(name, content string, activate bool) error {
	return nil
}

func (s *fakeSieveSession) Activate(name string) error {
	return nil
}

func (s *fakeSieveSession) Deactivate() error {
	return nil
}

func (s *fakeSieveSession) DeleteScript(name string) error {
	return nil
}

func (s *fakeSieveSession) RenameScript(oldName, newName string) error {
	return nil
}

func (s *fakeSieveSession) CheckScript(content string) error {
	return nil
}

func (s *fakeSieveSession) HaveSpace(name string, sizeBytes int) (bool, error) {
	return true, nil
}

func (s *fakeSieveSession) Close() {
	s.closed = true
}

type fakeSieveDialer struct {
	gotOpts SieveConnectOptions
	session *fakeSieveSession
}

func (d *fakeSieveDialer) DialSieve(_ context.Context, opts SieveConnectOptions) (SieveSession, error) {
	d.gotOpts = normalizeSieveConnectOptions(opts)
	if d.session == nil {
		d.session = &fakeSieveSession{}
	}
	return d.session, nil
}

type fakeStoredAccountResolver struct {
	gotAccountID string
	opts         ConnectOptions
}

func (r *fakeStoredAccountResolver) ResolveConnectOptions(_ context.Context, accountID string) (ConnectOptions, error) {
	r.gotAccountID = accountID
	return r.opts, nil
}

func TestBuildDSLRule(t *testing.T) {
	rule, err := BuildDSLRule(BuildRuleOptions{
		Name:             "invoice-search",
		SubjectContains:  "invoice",
		HasFlags:         []string{"seen"},
		Limit:            5,
		AfterUID:         10,
		IncludeContent:   true,
		ContentType:      "text/plain",
		ContentMaxLength: 200,
	})
	if err != nil {
		t.Fatalf("BuildDSLRule returned error: %v", err)
	}

	if rule.Name != "invoice-search" {
		t.Fatalf("rule.Name = %q, want %q", rule.Name, "invoice-search")
	}
	if rule.Search.SubjectContains != "invoice" {
		t.Fatalf("rule.Search.SubjectContains = %q, want invoice", rule.Search.SubjectContains)
	}
	if rule.Search.Flags == nil || len(rule.Search.Flags.Has) != 1 || rule.Search.Flags.Has[0] != "seen" {
		t.Fatalf("rule.Search.Flags = %#v, want has=[seen]", rule.Search.Flags)
	}
	if len(rule.Output.Fields) != 8 {
		t.Fatalf("len(rule.Output.Fields) = %d, want 8", len(rule.Output.Fields))
	}

	field, ok := rule.Output.Fields[7].(dsl.Field)
	if !ok {
		t.Fatalf("rule.Output.Fields[7] type = %T, want dsl.Field", rule.Output.Fields[7])
	}
	if field.Name != "mime_parts" {
		t.Fatalf("field.Name = %q, want mime_parts", field.Name)
	}
	if field.Content == nil || field.Content.Mode != "filter" {
		t.Fatalf("field.Content = %#v, want filter mode", field.Content)
	}
	if len(field.Content.Types) != 1 || field.Content.Types[0] != "text/plain" {
		t.Fatalf("field.Content.Types = %#v, want [text/plain]", field.Content.Types)
	}
}

func TestParseRuleMap(t *testing.T) {
	service := New()
	ruleMap, err := service.ParseRuleMap(`
name: export-rule
description: export invoices
search:
  subject_contains: invoice
output:
  format: json
  fields:
    - uid
    - subject
actions:
  export:
    format: eml
`)
	if err != nil {
		t.Fatalf("ParseRuleMap returned error: %v", err)
	}

	if got := ruleMap["name"]; got != "export-rule" {
		t.Fatalf("ruleMap[name] = %#v, want export-rule", got)
	}
	search, ok := ruleMap["search"].(map[string]interface{})
	if !ok {
		t.Fatalf("ruleMap[search] type = %T, want map[string]interface{}", ruleMap["search"])
	}
	if got := search["subjectContains"]; got != "invoice" {
		t.Fatalf("search[subjectContains] = %#v, want invoice", got)
	}
	actions, ok := ruleMap["actions"].(map[string]interface{})
	if !ok {
		t.Fatalf("ruleMap[actions] type = %T, want map[string]interface{}", ruleMap["actions"])
	}
	if _, ok := actions["export"]; !ok {
		t.Fatalf("actions = %#v, want export key", actions)
	}
}

func TestShapeMessageMap(t *testing.T) {
	service := New()
	messageMap, err := service.ShapeMessageMap(&dsl.EmailMessage{
		UID:        42,
		SeqNum:     7,
		Flags:      []string{"seen"},
		Size:       1234,
		TotalCount: 9,
		Envelope: &dsl.EmailEnvelope{
			Subject: "Hello",
			Date:    time.Date(2026, time.March, 8, 12, 0, 0, 0, time.UTC),
			From: []dsl.EmailAddress{{
				Name:    "Sender",
				Address: "sender@example.com",
			}},
			To: []dsl.EmailAddress{{
				Name:    "Recipient",
				Address: "recipient@example.com",
			}},
		},
		MimeParts: []dsl.MimePart{{
			Type:    "text/plain",
			Content: "hello world",
		}},
	})
	if err != nil {
		t.Fatalf("ShapeMessageMap returned error: %v", err)
	}

	if got := messageMap["uid"]; got != float64(42) {
		t.Fatalf("messageMap[uid] = %#v, want 42", got)
	}
	if got := messageMap["subject"]; got != "Hello" {
		t.Fatalf("messageMap[subject] = %#v, want Hello", got)
	}
}

func TestConnectUsesInjectedDialer(t *testing.T) {
	dialer := &fakeDialer{}
	service := New(WithDialer(dialer))

	session, err := service.Connect(context.Background(), ConnectOptions{
		Server:   "imap.example.com",
		Username: "user@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	if session.Mailbox() != "INBOX" {
		t.Fatalf("session.Mailbox() = %q, want INBOX", session.Mailbox())
	}
	if dialer.gotOpts.Port != 993 {
		t.Fatalf("dialer.gotOpts.Port = %d, want 993", dialer.gotOpts.Port)
	}
	if dialer.gotOpts.Mailbox != "INBOX" {
		t.Fatalf("dialer.gotOpts.Mailbox = %q, want INBOX", dialer.gotOpts.Mailbox)
	}
	session.Close()
	if !dialer.session.closed {
		t.Fatalf("session was not closed")
	}
}

func TestConnectUsesStoredAccountResolver(t *testing.T) {
	dialer := &fakeDialer{}
	resolver := &fakeStoredAccountResolver{
		opts: ConnectOptions{
			Server:   "imap.example.com",
			Port:     993,
			Username: "user@example.com",
			Password: "secret",
			Mailbox:  "Archive",
		},
	}
	service := New(WithDialer(dialer), WithStoredAccountResolver(resolver))

	session, err := service.Connect(context.Background(), ConnectOptions{
		AccountID: "acc-1",
	})
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	if resolver.gotAccountID != "acc-1" {
		t.Fatalf("resolver.gotAccountID = %q, want acc-1", resolver.gotAccountID)
	}
	if dialer.gotOpts.Username != "user@example.com" || dialer.gotOpts.Password != "secret" {
		t.Fatalf("unexpected dialer opts: %+v", dialer.gotOpts)
	}
	if session.Mailbox() != "Archive" {
		t.Fatalf("session.Mailbox() = %q, want Archive", session.Mailbox())
	}
}

func TestConnectSieveUsesInjectedDialer(t *testing.T) {
	dialer := &fakeSieveDialer{}
	service := New(WithSieveDialer(dialer))

	session, err := service.ConnectSieve(context.Background(), SieveConnectOptions{
		Server:   "sieve.example.com",
		Username: "user@example.com",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("ConnectSieve returned error: %v", err)
	}

	if dialer.gotOpts.Port != 4190 {
		t.Fatalf("dialer.gotOpts.Port = %d, want 4190", dialer.gotOpts.Port)
	}
	session.Close()
	if !dialer.session.closed {
		t.Fatalf("expected sieve session to be closed")
	}
}

func TestConnectSieveUsesStoredAccountResolver(t *testing.T) {
	sieveDialer := &fakeSieveDialer{}
	resolver := &fakeStoredAccountResolver{
		opts: ConnectOptions{
			Server:   "imap.example.com",
			Port:     993,
			Username: "user@example.com",
			Password: "secret",
			Mailbox:  "Archive",
		},
	}
	service := New(WithSieveDialer(sieveDialer), WithStoredAccountResolver(resolver))

	_, err := service.ConnectSieve(context.Background(), SieveConnectOptions{
		AccountID: "acc-1",
	})
	if err != nil {
		t.Fatalf("ConnectSieve returned error: %v", err)
	}

	if resolver.gotAccountID != "acc-1" {
		t.Fatalf("resolver.gotAccountID = %q, want acc-1", resolver.gotAccountID)
	}
	if sieveDialer.gotOpts.Server != "imap.example.com" {
		t.Fatalf("sieveDialer.gotOpts.Server = %q, want imap.example.com", sieveDialer.gotOpts.Server)
	}
	if sieveDialer.gotOpts.Username != "user@example.com" || sieveDialer.gotOpts.Password != "secret" {
		t.Fatalf("unexpected sieve dialer opts: %+v", sieveDialer.gotOpts)
	}
	if sieveDialer.gotOpts.Port != 4190 {
		t.Fatalf("sieveDialer.gotOpts.Port = %d, want 4190", sieveDialer.gotOpts.Port)
	}
}
