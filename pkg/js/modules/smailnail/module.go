package smailnailmodule

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	ggjmodules "github.com/go-go-golems/go-go-goja/modules"
	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
)

type Module struct {
	service *smailnailjs.Service
}

var _ ggjmodules.NativeModule = (*Module)(nil)

func NewModule() *Module {
	return NewModuleWithService(smailnailjs.New())
}

func NewModuleWithService(service *smailnailjs.Service) *Module {
	if service == nil {
		service = smailnailjs.New()
	}
	return &Module{service: service}
}

func init() {
	ggjmodules.Register(NewModule())
}

func (m *Module) Name() string {
	return "smailnail"
}

func (m *Module) Doc() string {
	return "smailnail exposes rule helpers and IMAP service construction for JavaScript runtimes."
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
	ggjmodules.SetExport(obj, m.Name(), "connect", func(input map[string]interface{}) (*goja.Object, error) {
		opts, err := smailnailjs.DecodeConnectOptions(input)
		if err != nil {
			return nil, err
		}
		session, err := m.service.Connect(context.Background(), opts)
		if err != nil {
			return nil, err
		}
		return m.newSessionObject(rt, session)
	})

	return obj
}

func (m *Module) newSessionObject(rt *goja.Runtime, session smailnailjs.Session) (*goja.Object, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}
	obj := rt.NewObject()
	if err := obj.Set("mailbox", session.Mailbox()); err != nil {
		return nil, err
	}
	ggjmodules.SetExport(obj, m.Name(), "close", func() {
		session.Close()
	})
	return obj, nil
}
