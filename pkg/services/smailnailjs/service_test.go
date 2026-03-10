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
