package smailnailmodule

import (
	"context"
	"testing"

	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
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
	session *fakeSession
}

func (d *fakeDialer) Dial(_ context.Context, opts smailnailjs.ConnectOptions) (smailnailjs.Session, error) {
	mailbox := opts.Mailbox
	if mailbox == "" {
		mailbox = "INBOX"
	}
	if d.session == nil {
		d.session = &fakeSession{mailbox: mailbox}
	}
	return d.session, nil
}

func TestModuleBuildRule(t *testing.T) {
	module := NewModule()
	factory, err := ggjengine.NewBuilder().
		WithModules(ggjengine.NativeModuleSpec{
			ModuleName: module.Name(),
			Loader:     module.Loader,
		}).
		Build()
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}
	defer func() {
		_ = rt.Close(context.Background())
	}()

	value, err := rt.VM.RunString(`
		const smailnail = require("smailnail");
		const rule = smailnail.buildRule({
			name: "invoice-search",
			subjectContains: "invoice",
			includeContent: true,
			contentType: "text/plain"
		});
		rule.search.subjectContains + "|" + rule.output.fields[7].content.types[0];
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	if got := value.String(); got != "invoice|text/plain" {
		t.Fatalf("result = %q, want %q", got, "invoice|text/plain")
	}
}

func TestModuleNewServiceConnect(t *testing.T) {
	dialer := &fakeDialer{}
	module := NewModuleWithService(smailnailjs.New(smailnailjs.WithDialer(dialer)))
	factory, err := ggjengine.NewBuilder().
		WithModules(ggjengine.NativeModuleSpec{
			ModuleName: module.Name(),
			Loader:     module.Loader,
		}).
		Build()
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("NewRuntime returned error: %v", err)
	}
	defer func() {
		_ = rt.Close(context.Background())
	}()

	value, err := rt.VM.RunString(`
		const smailnail = require("smailnail");
		const svc = smailnail.newService();
		const session = svc.connect({
			server: "imap.example.com",
			username: "user@example.com",
			password: "secret"
		});
		session.close();
		session.mailbox;
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	if got := value.String(); got != "INBOX" {
		t.Fatalf("result = %q, want %q", got, "INBOX")
	}
	if dialer.session == nil || !dialer.session.closed {
		t.Fatalf("expected fake session to be closed")
	}
}
