package smailnailmodule

import (
	"context"
	"encoding/json"
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"time"

	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/extract"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/model"
	smailnaildocs "github.com/go-go-golems/smailnail/pkg/js/modules/smailnail/docs"
	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
)

type fakeSession struct {
	mailbox      string
	closed       bool
	lastMovedTo  string
	lastCopiedTo string
}

func (s *fakeSession) Mailbox() string {
	return s.mailbox
}

func (s *fakeSession) Capabilities() map[string]bool {
	return map[string]bool{"uidplus": true, "move": true}
}

func (s *fakeSession) List(pattern string) ([]smailnailjs.MailboxInfo, error) {
	return []smailnailjs.MailboxInfo{
		{Name: "INBOX", Delimiter: "/"},
		{Name: "Archive", Delimiter: "/"},
	}, nil
}

func (s *fakeSession) Status(name string) (*smailnailjs.MailboxStatus, error) {
	return &smailnailjs.MailboxStatus{Messages: 7, UIDNext: 8}, nil
}

func (s *fakeSession) SelectMailbox(name string, readOnly bool) (*smailnailjs.MailboxSelection, error) {
	s.mailbox = name
	return &smailnailjs.MailboxSelection{Name: name, ReadOnly: readOnly}, nil
}

func (s *fakeSession) Search(criteria *smailnailjs.SearchCriteria) ([]uint32, error) {
	return []uint32{101, 202}, nil
}

func (s *fakeSession) Fetch(uids []uint32, fields []smailnailjs.FetchField) ([]*smailnailjs.FetchedMessage, error) {
	return []*smailnailjs.FetchedMessage{
		{UID: uids[0], Flags: []string{"\\Seen"}, BodyText: "hello"},
	}, nil
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
	s.lastMovedTo = dest
	return nil
}

func (s *fakeSession) Copy(uids []uint32, dest string) error {
	s.lastCopiedTo = dest
	return nil
}

func (s *fakeSession) Delete(uids []uint32, expunge bool) error {
	return nil
}

func (s *fakeSession) Expunge(uids []uint32) error {
	return nil
}

func (s *fakeSession) Append(mailbox string, message []byte, flags []string, date *time.Time) (uint32, error) {
	return 303, nil
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

type fakeStoredAccountResolver struct {
	gotAccountID string
	opts         smailnailjs.ConnectOptions
}

func (r *fakeStoredAccountResolver) ResolveConnectOptions(_ context.Context, accountID string) (smailnailjs.ConnectOptions, error) {
	r.gotAccountID = accountID
	return r.opts, nil
}

type fakeSieveSession struct {
	closed bool
}

func (s *fakeSieveSession) Capabilities() smailnailjs.SieveCapabilities {
	return smailnailjs.SieveCapabilities{Implementation: "FakeSieve", Sieve: []string{"fileinto"}}
}

func (s *fakeSieveSession) ListScripts() ([]smailnailjs.ScriptInfo, error) {
	return []smailnailjs.ScriptInfo{{Name: "main", Active: true}}, nil
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
	session *fakeSieveSession
}

func (d *fakeSieveDialer) DialSieve(_ context.Context, opts smailnailjs.SieveConnectOptions) (smailnailjs.SieveSession, error) {
	if d.session == nil {
		d.session = &fakeSieveSession{}
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

func TestModuleIMAPSessionOperations(t *testing.T) {
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
		const service = smailnail.newService();
		const session = service.connect({
			server: "imap.example.com",
			username: "user@example.com",
			password: "secret"
		});
		const before = session.mailbox;
		const boxes = session.list();
		const status = session.status("INBOX");
		const selection = session.selectMailbox("Archive", { readOnly: true });
		const uids = session.search({ subject: "invoice" });
		const messages = session.fetch(uids, ["uid", "flags", "body.text"]);
		session.addFlags(uids, ["\\\\Seen"]);
		session.removeFlags(uids, ["\\\\Flagged"]);
		session.setFlags(uids, ["\\\\Seen", "\\\\Flagged"], { silent: true });
		session.move(uids, "Processed");
		session.copy(uids, "Backup");
		session.delete(uids, { expunge: false });
		session.expunge();
		const appended = session.append("Subject: Hello\\r\\n\\r\\nbody", { flags: ["\\\\Seen"] });
		const capabilities = session.capabilities();
		session.close();
		JSON.stringify({
			before,
			after: session.mailbox,
			boxCount: boxes.length,
			statusMessages: status.messages,
			selected: selection.name,
			firstUID: uids[0],
			firstFetchedUID: messages[0].uid,
			appended,
			uidplus: capabilities.uidplus === true,
		});
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	var decoded struct {
		Before         string `json:"before"`
		After          string `json:"after"`
		BoxCount       int    `json:"boxCount"`
		StatusMessages uint32 `json:"statusMessages"`
		Selected       string `json:"selected"`
		FirstUID       uint32 `json:"firstUID"`
		FirstFetched   uint32 `json:"firstFetchedUID"`
		Appended       uint32 `json:"appended"`
		UIDPlus        bool   `json:"uidplus"`
	}
	if err := json.Unmarshal([]byte(value.String()), &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if decoded.Before != "INBOX" || decoded.After != "Archive" {
		t.Fatalf("unexpected mailbox transition: %+v", decoded)
	}
	if decoded.BoxCount != 2 || decoded.StatusMessages != 7 {
		t.Fatalf("unexpected mailbox metadata: %+v", decoded)
	}
	if decoded.Selected != "Archive" || decoded.FirstUID != 101 || decoded.FirstFetched != 101 {
		t.Fatalf("unexpected query results: %+v", decoded)
	}
	if decoded.Appended != 303 || !decoded.UIDPlus {
		t.Fatalf("unexpected append/capability result: %+v", decoded)
	}
	if dialer.session == nil || !dialer.session.closed {
		t.Fatalf("expected fake session to be closed")
	}
}

func TestModuleBuildSieveScriptAndConnectSieve(t *testing.T) {
	module := NewModuleWithService(smailnailjs.New(smailnailjs.WithSieveDialer(&fakeSieveDialer{})))
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
		const service = smailnail.newService();
		const script = service.buildSieveScript((s) => {
			s.require(["fileinto"]);
			s.if(s.headerContains("Subject", "invoice"), (a) => {
				a.fileInto("Invoices");
				a.stop();
			});
		});
		const sieve = service.connectSieve({
			server: "sieve.example.com",
			username: "user@example.com",
			password: "secret"
		});
		const caps = sieve.capabilities();
		const scripts = sieve.listScripts();
		const current = sieve.getScript("main");
		sieve.putScript("main", script, { activate: true });
		sieve.activate("main");
		sieve.renameScript("main", "main-renamed");
		sieve.deleteScript("main-renamed");
		sieve.deactivate();
		sieve.check(script);
		const space = sieve.haveSpace("main", script.length);
		sieve.close();
		JSON.stringify({
			script,
			implementation: caps.implementation,
			scriptCount: scripts.length,
			current,
			space,
		});
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	var decoded struct {
		Script         string `json:"script"`
		Implementation string `json:"implementation"`
		ScriptCount    int    `json:"scriptCount"`
		Current        string `json:"current"`
		Space          bool   `json:"space"`
	}
	if err := json.Unmarshal([]byte(value.String()), &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if decoded.Implementation != "FakeSieve" || decoded.ScriptCount != 1 || !decoded.Space {
		t.Fatalf("unexpected sieve metadata: %+v", decoded)
	}
	if decoded.Current != "keep;" {
		t.Fatalf("decoded.Current = %q, want keep;", decoded.Current)
	}
	if want := "require [\"fileinto\"];\nif header :contains \"Subject\" \"invoice\" {\n  fileinto \"Invoices\";\n  stop;\n}\n"; decoded.Script != want {
		t.Fatalf("decoded.Script = %q, want %q", decoded.Script, want)
	}
}

func TestModuleNewServiceConnectWithStoredAccount(t *testing.T) {
	dialer := &fakeDialer{}
	resolver := &fakeStoredAccountResolver{
		opts: smailnailjs.ConnectOptions{
			Server:   "imap.example.com",
			Username: "user@example.com",
			Password: "secret",
			Mailbox:  "Archive",
		},
	}
	module := NewModuleWithService(smailnailjs.New(
		smailnailjs.WithDialer(dialer),
		smailnailjs.WithStoredAccountResolver(resolver),
	))
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
		const session = svc.connect({ accountId: "acc-1" });
		session.mailbox;
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	if got := value.String(); got != "Archive" {
		t.Fatalf("result = %q, want %q", got, "Archive")
	}
	if resolver.gotAccountID != "acc-1" {
		t.Fatalf("resolver.gotAccountID = %q, want acc-1", resolver.gotAccountID)
	}
}

func TestDocumentedSymbolsMatchRuntimeExports(t *testing.T) {
	store := loadEmbeddedDocStore(t)
	module := NewModuleWithService(smailnailjs.New(
		smailnailjs.WithDialer(&fakeDialer{}),
		smailnailjs.WithSieveDialer(&fakeSieveDialer{}),
	))
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
		function callableKeys(obj) {
			return Object.keys(obj).filter((k) => typeof obj[k] === "function").sort();
		}
		const smailnail = require("smailnail");
		const service = smailnail.newService();
		const imap = service.connect({
			server: "imap.example.com",
			username: "user@example.com",
			password: "secret"
		});
		const sieve = service.connectSieve({
			server: "sieve.example.com",
			username: "user@example.com",
			password: "secret"
		});
		const result = {
			topLevel: callableKeys(smailnail),
			service: callableKeys(service),
			imap: callableKeys(imap),
			sieve: callableKeys(sieve),
		};
		imap.close();
		sieve.close();
		JSON.stringify(result);
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	var decoded struct {
		TopLevel []string `json:"topLevel"`
		Service  []string `json:"service"`
		IMAP     []string `json:"imap"`
		Sieve    []string `json:"sieve"`
	}
	if err := json.Unmarshal([]byte(value.String()), &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	runtimeSymbols := uniqueSorted(append(append(append(decoded.TopLevel, decoded.Service...), decoded.IMAP...), decoded.Sieve...))
	documentedSymbols := make([]string, 0, len(store.BySymbol))
	for name := range store.BySymbol {
		documentedSymbols = append(documentedSymbols, name)
	}
	sort.Strings(documentedSymbols)

	if !reflect.DeepEqual(runtimeSymbols, documentedSymbols) {
		t.Fatalf("runtime symbols = %v, documented symbols = %v", runtimeSymbols, documentedSymbols)
	}

	for id, example := range store.ByExample {
		for _, name := range example.Symbols {
			if store.BySymbol[name] == nil {
				t.Fatalf("example %q references undocumented symbol %q", id, name)
			}
		}
	}
}

func loadEmbeddedDocStore(t *testing.T) *model.DocStore {
	t.Helper()

	entries, err := fs.ReadDir(smailnaildocs.Files, smailnaildocs.Dir)
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}

	store := model.NewDocStore()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fd, err := extract.ParseFSFile(smailnaildocs.Files, entry.Name())
		if err != nil {
			t.Fatalf("ParseFSFile(%s) returned error: %v", entry.Name(), err)
		}
		store.AddFile(fd)
	}
	return store
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	ret := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ret = append(ret, value)
	}
	sort.Strings(ret)
	return ret
}
