package smailnailmodule

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
	ggjmodules "github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
)

type Module struct {
	ctx     context.Context
	service *smailnailjs.Service
}

var _ ggjmodules.NativeModule = (*Module)(nil)

func NewModule() *Module {
	return NewModuleWithServiceAndContext(context.Background(), smailnailjs.New())
}

func NewModuleWithService(service *smailnailjs.Service) *Module {
	return NewModuleWithServiceAndContext(context.Background(), service)
}

func NewModuleWithServiceAndContext(ctx context.Context, service *smailnailjs.Service) *Module {
	if service == nil {
		service = smailnailjs.New()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return &Module{ctx: ctx, service: service}
}

func init() {
	ggjmodules.Register(NewModule())
}

func (m *Module) Name() string {
	return "smailnail"
}

func (m *Module) Doc() string {
	return "smailnail exposes rule helpers, IMAP automation, and ManageSieve scripting for JavaScript runtimes."
}

func (m *Module) Loader(rt *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	ggjmodules.SetExport(exports, m.Name(), "parseRule", func(yamlString string) (map[string]interface{}, error) {
		return m.service.ParseRuleMap(yamlString)
	})
	ggjmodules.SetExport(exports, m.Name(), "buildRule", func(input map[string]interface{}) (map[string]interface{}, error) {
		opts, err := smailnailjs.DecodeBuildRuleOptions(input)
		if err != nil {
			return nil, err
		}
		return m.service.BuildRuleMap(opts)
	})
	if err := exports.Set("buildSieveScript", func(call goja.FunctionCall) goja.Value {
		return buildSieveScript(rt, call)
	}); err != nil {
		panic(rt.NewGoError(err))
	}
	ggjmodules.SetExport(exports, m.Name(), "newService", func() *goja.Object {
		return m.newServiceObject(rt)
	})
}

func (m *Module) newServiceObject(rt *goja.Runtime) *goja.Object {
	obj := rt.NewObject()

	ggjmodules.SetExport(obj, m.Name(), "parseRule", func(yamlString string) (map[string]interface{}, error) {
		return m.service.ParseRuleMap(yamlString)
	})
	ggjmodules.SetExport(obj, m.Name(), "buildRule", func(input map[string]interface{}) (map[string]interface{}, error) {
		opts, err := smailnailjs.DecodeBuildRuleOptions(input)
		if err != nil {
			return nil, err
		}
		return m.service.BuildRuleMap(opts)
	})
	if err := obj.Set("buildSieveScript", func(call goja.FunctionCall) goja.Value {
		return buildSieveScript(rt, call)
	}); err != nil {
		panic(rt.NewGoError(err))
	}
	ggjmodules.SetExport(obj, m.Name(), "connect", func(input map[string]interface{}) (*goja.Object, error) {
		opts, err := smailnailjs.DecodeConnectOptions(input)
		if err != nil {
			return nil, err
		}
		session, err := m.service.Connect(m.ctx, opts)
		if err != nil {
			return nil, err
		}
		return m.newIMAPSessionObject(rt, session)
	})
	ggjmodules.SetExport(obj, m.Name(), "connectSieve", func(input map[string]interface{}) (*goja.Object, error) {
		opts, err := smailnailjs.DecodeSieveConnectOptions(input)
		if err != nil {
			return nil, err
		}
		session, err := m.service.ConnectSieve(m.ctx, opts)
		if err != nil {
			return nil, err
		}
		return m.newSieveSessionObject(rt, session)
	})

	return obj
}

func (m *Module) newIMAPSessionObject(rt *goja.Runtime, session smailnailjs.Session) (*goja.Object, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}
	obj := rt.NewObject()
	if err := obj.Set("mailbox", session.Mailbox()); err != nil {
		return nil, err
	}
	if err := obj.Set("capabilities", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue(session.Capabilities())
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("list", func(call goja.FunctionCall) goja.Value {
		pattern := ""
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) {
			pattern = call.Arguments[0].String()
		}
		boxes, err := session.List(pattern)
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return toJSONValue(rt, boxes)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("status", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		status, err := session.Status(name)
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return toJSONValue(rt, status)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("selectMailbox", func(call goja.FunctionCall) goja.Value {
		name := call.Argument(0).String()
		readOnly := false
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
			options := call.Arguments[1].ToObject(rt)
			if v := options.Get("readOnly"); v != nil && !goja.IsUndefined(v) {
				readOnly = v.ToBoolean()
			}
		}
		selection, err := session.SelectMailbox(name, readOnly)
		if err != nil {
			panic(rt.NewGoError(err))
		}
		if err := obj.Set("mailbox", session.Mailbox()); err != nil {
			panic(rt.NewGoError(err))
		}
		return toJSONValue(rt, selection)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("search", func(call goja.FunctionCall) goja.Value {
		var criteria *smailnailjs.SearchCriteria
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) {
			criteria = parseCriteria(rt, call.Arguments[0])
		}
		uids, err := session.Search(criteria)
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return rt.ToValue(uids)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("fetch", func(call goja.FunctionCall) goja.Value {
		uids := jsToUint32Slice(call.Argument(0))
		fields := jsToFetchFields(call.Argument(1))
		msgs, err := session.Fetch(uids, fields)
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return toJSONValue(rt, msgs)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("addFlags", func(call goja.FunctionCall) goja.Value {
		if err := session.AddFlags(jsToUint32Slice(call.Argument(0)), jsToStringSlice(call.Argument(1)), readSilentOption(rt, call, 2)); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("removeFlags", func(call goja.FunctionCall) goja.Value {
		if err := session.RemoveFlags(jsToUint32Slice(call.Argument(0)), jsToStringSlice(call.Argument(1)), readSilentOption(rt, call, 2)); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("setFlags", func(call goja.FunctionCall) goja.Value {
		if err := session.SetFlags(jsToUint32Slice(call.Argument(0)), jsToStringSlice(call.Argument(1)), readSilentOption(rt, call, 2)); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("move", func(call goja.FunctionCall) goja.Value {
		if err := session.Move(jsToUint32Slice(call.Argument(0)), call.Argument(1).String()); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("copy", func(call goja.FunctionCall) goja.Value {
		if err := session.Copy(jsToUint32Slice(call.Argument(0)), call.Argument(1).String()); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("delete", func(call goja.FunctionCall) goja.Value {
		expunge := true
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
			options := call.Arguments[1].ToObject(rt)
			if v := options.Get("expunge"); v != nil && !goja.IsUndefined(v) {
				expunge = v.ToBoolean()
			}
		}
		if err := session.Delete(jsToUint32Slice(call.Argument(0)), expunge); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("expunge", func(call goja.FunctionCall) goja.Value {
		var uids []uint32
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Arguments[0]) && !goja.IsNull(call.Arguments[0]) {
			uids = jsToUint32Slice(call.Arguments[0])
		}
		if err := session.Expunge(uids); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("append", func(call goja.FunctionCall) goja.Value {
		content := call.Argument(0).String()
		mailbox := session.Mailbox()
		var flags []string
		var date *time.Time
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
			options := call.Arguments[1].ToObject(rt)
			if v := options.Get("mailbox"); v != nil && !goja.IsUndefined(v) {
				mailbox = v.String()
			}
			if v := options.Get("flags"); v != nil && !goja.IsUndefined(v) {
				flags = jsToStringSlice(v)
			}
			if v := options.Get("date"); v != nil && !goja.IsUndefined(v) {
				t := parseJSDate(v)
				if !t.IsZero() {
					date = &t
				}
			}
		}
		uid, err := session.Append(mailbox, []byte(content), flags, date)
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return rt.ToValue(uid)
	}); err != nil {
		return nil, err
	}
	ggjmodules.SetExport(obj, m.Name(), "close", func() {
		session.Close()
	})
	return obj, nil
}

func (m *Module) newSieveSessionObject(rt *goja.Runtime, session smailnailjs.SieveSession) (*goja.Object, error) {
	if session == nil {
		return nil, fmt.Errorf("sieve session is nil")
	}
	obj := rt.NewObject()
	if err := obj.Set("capabilities", func(call goja.FunctionCall) goja.Value {
		return toJSONValue(rt, session.Capabilities())
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("listScripts", func(call goja.FunctionCall) goja.Value {
		scripts, err := session.ListScripts()
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return toJSONValue(rt, scripts)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("getScript", func(call goja.FunctionCall) goja.Value {
		content, err := session.GetScript(call.Argument(0).String())
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return rt.ToValue(content)
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("putScript", func(call goja.FunctionCall) goja.Value {
		activate := false
		if len(call.Arguments) > 2 && !goja.IsUndefined(call.Arguments[2]) && !goja.IsNull(call.Arguments[2]) {
			options := call.Arguments[2].ToObject(rt)
			if v := options.Get("activate"); v != nil && !goja.IsUndefined(v) {
				activate = v.ToBoolean()
			}
		}
		if err := session.PutScript(call.Argument(0).String(), call.Argument(1).String(), activate); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("activate", func(call goja.FunctionCall) goja.Value {
		if err := session.Activate(call.Argument(0).String()); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("deactivate", func(call goja.FunctionCall) goja.Value {
		if err := session.Deactivate(); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("deleteScript", func(call goja.FunctionCall) goja.Value {
		if err := session.DeleteScript(call.Argument(0).String()); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("renameScript", func(call goja.FunctionCall) goja.Value {
		if err := session.RenameScript(call.Argument(0).String(), call.Argument(1).String()); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("check", func(call goja.FunctionCall) goja.Value {
		if err := session.CheckScript(call.Argument(0).String()); err != nil {
			panic(rt.NewGoError(err))
		}
		return goja.Undefined()
	}); err != nil {
		return nil, err
	}
	if err := obj.Set("haveSpace", func(call goja.FunctionCall) goja.Value {
		ok, err := session.HaveSpace(call.Argument(0).String(), int(call.Argument(1).ToInteger()))
		if err != nil {
			panic(rt.NewGoError(err))
		}
		return rt.ToValue(ok)
	}); err != nil {
		return nil, err
	}
	ggjmodules.SetExport(obj, m.Name(), "close", func() {
		session.Close()
	})
	return obj, nil
}
