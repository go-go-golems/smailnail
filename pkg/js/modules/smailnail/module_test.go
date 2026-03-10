package smailnailmodule

import (
	"context"
	"encoding/json"
	"io/fs"
	"reflect"
	"sort"
	"testing"

	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/extract"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/model"
	smailnaildocs "github.com/go-go-golems/smailnail/pkg/js/modules/smailnail/docs"
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

func TestDocumentedSymbolsMatchRuntimeExports(t *testing.T) {
	store := loadEmbeddedDocStore(t)
	module := NewModuleWithService(smailnailjs.New(smailnailjs.WithDialer(&fakeDialer{})))
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
		const session = service.connect({
			server: "imap.example.com",
			username: "user@example.com",
			password: "secret"
		});
		const result = {
			topLevel: callableKeys(smailnail),
			service: callableKeys(service),
			session: callableKeys(session),
		};
		session.close();
		JSON.stringify(result);
	`)
	if err != nil {
		t.Fatalf("RunString returned error: %v", err)
	}

	var decoded struct {
		TopLevel []string `json:"topLevel"`
		Service  []string `json:"service"`
		Session  []string `json:"session"`
	}
	if err := json.Unmarshal([]byte(value.String()), &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	runtimeSymbols := uniqueSorted(append(append(decoded.TopLevel, decoded.Service...), decoded.Session...))
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
