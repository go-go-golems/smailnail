package smailnailmodule

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/smailnail/pkg/mailruntime"
)

func buildSieveScript(rt *goja.Runtime, call goja.FunctionCall) goja.Value {
	fn, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		panic(rt.NewTypeError("buildSieveScript: argument must be a function"))
	}
	builder := newSieveBuilder(rt)
	if _, err := fn(goja.Undefined(), builder.obj); err != nil {
		panic(rt.NewGoError(err))
	}
	return rt.ToValue(builder.String())
}

type sieveBuilder struct {
	rt     *goja.Runtime
	obj    *goja.Object
	lines  []string
	indent int
}

type sieveActionBuilder struct {
	rt     *goja.Runtime
	obj    *goja.Object
	parent *sieveBuilder
}

func newSieveBuilder(rt *goja.Runtime) *sieveBuilder {
	b := &sieveBuilder{rt: rt}
	b.obj = rt.NewObject()

	_ = b.obj.Set("require", func(call goja.FunctionCall) goja.Value {
		exts := jsToStringSlice(call.Argument(0))
		quoted := make([]string, len(exts))
		for i, ext := range exts {
			quoted[i] = fmt.Sprintf("%q", ext)
		}
		b.lines = append(b.lines, fmt.Sprintf("require [%s];", strings.Join(quoted, ", ")))
		return goja.Undefined()
	})

	_ = b.obj.Set("if", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(rt.NewTypeError("sieve.if(condition, actionFn) requires 2 arguments"))
		}
		cond := call.Arguments[0].String()
		actionFn, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			panic(rt.NewTypeError("sieve.if: second argument must be a function"))
		}
		b.lines = append(b.lines, fmt.Sprintf("if %s {", cond))
		b.indent++
		ab := newSieveActionBuilder(rt, b)
		if _, err := actionFn(goja.Undefined(), ab.obj); err != nil {
			panic(rt.NewGoError(err))
		}
		b.indent--
		b.lines = append(b.lines, "}")
		return goja.Undefined()
	})

	_ = b.obj.Set("all", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		return rt.ToValue(fmt.Sprintf("allof(%s)", strings.Join(parts, ", ")))
	})

	_ = b.obj.Set("any", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		return rt.ToValue(fmt.Sprintf("anyof(%s)", strings.Join(parts, ", ")))
	})

	_ = b.obj.Set("not", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue(fmt.Sprintf("not %s", call.Argument(0).String()))
	})

	_ = b.obj.Set("headerContains", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue(fmt.Sprintf("header :contains %q %q", call.Argument(0).String(), call.Argument(1).String()))
	})

	_ = b.obj.Set("headerMatches", func(call goja.FunctionCall) goja.Value {
		pattern := mailruntime.StringOrRegex(call.Argument(1))
		return rt.ToValue(fmt.Sprintf("header :matches %q %q", call.Argument(0).String(), pattern))
	})

	_ = b.obj.Set("headerIs", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue(fmt.Sprintf("header :is %q %q", call.Argument(0).String(), call.Argument(1).String()))
	})

	_ = b.obj.Set("address", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue(fmt.Sprintf("address :%s %q %q", call.Argument(0).String(), call.Argument(1).String(), call.Argument(2).String()))
	})

	_ = b.obj.Set("size", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue(fmt.Sprintf("size :%s %d", call.Argument(0).String(), call.Argument(1).ToInteger()))
	})

	_ = b.obj.Set("true", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue("true")
	})

	_ = b.obj.Set("false", func(call goja.FunctionCall) goja.Value {
		return rt.ToValue("false")
	})

	return b
}

func newSieveActionBuilder(rt *goja.Runtime, parent *sieveBuilder) *sieveActionBuilder {
	ab := &sieveActionBuilder{rt: rt, parent: parent}
	ab.obj = rt.NewObject()

	addLine := func(line string) {
		parent.lines = append(parent.lines, strings.Repeat("  ", parent.indent)+line)
	}

	_ = ab.obj.Set("fileInto", func(call goja.FunctionCall) goja.Value {
		addLine(fmt.Sprintf("fileinto %q;", call.Argument(0).String()))
		return goja.Undefined()
	})
	_ = ab.obj.Set("redirect", func(call goja.FunctionCall) goja.Value {
		addLine(fmt.Sprintf("redirect %q;", call.Argument(0).String()))
		return goja.Undefined()
	})
	_ = ab.obj.Set("keep", func(call goja.FunctionCall) goja.Value {
		addLine("keep;")
		return goja.Undefined()
	})
	_ = ab.obj.Set("discard", func(call goja.FunctionCall) goja.Value {
		addLine("discard;")
		return goja.Undefined()
	})
	_ = ab.obj.Set("stop", func(call goja.FunctionCall) goja.Value {
		addLine("stop;")
		return goja.Undefined()
	})
	_ = ab.obj.Set("addFlag", func(call goja.FunctionCall) goja.Value {
		addLine(fmt.Sprintf("addflag %q;", call.Argument(0).String()))
		return goja.Undefined()
	})
	_ = ab.obj.Set("removeFlag", func(call goja.FunctionCall) goja.Value {
		addLine(fmt.Sprintf("removeflag %q;", call.Argument(0).String()))
		return goja.Undefined()
	})
	_ = ab.obj.Set("setFlag", func(call goja.FunctionCall) goja.Value {
		addLine(fmt.Sprintf("setflag %q;", call.Argument(0).String()))
		return goja.Undefined()
	})
	_ = ab.obj.Set("setVariable", func(call goja.FunctionCall) goja.Value {
		addLine(fmt.Sprintf("set %q %q;", call.Argument(0).String(), call.Argument(1).String()))
		return goja.Undefined()
	})
	_ = ab.obj.Set("vacation", func(call goja.FunctionCall) goja.Value {
		options := call.Argument(0).ToObject(rt)
		var parts []string
		if v := options.Get("days"); v != nil && !goja.IsUndefined(v) {
			parts = append(parts, fmt.Sprintf(":days %d", v.ToInteger()))
		}
		if v := options.Get("subject"); v != nil && !goja.IsUndefined(v) {
			parts = append(parts, fmt.Sprintf(":subject %q", v.String()))
		}
		message := ""
		if v := options.Get("message"); v != nil && !goja.IsUndefined(v) {
			message = v.String()
		}
		addLine(fmt.Sprintf("vacation %s %q;", strings.Join(parts, " "), message))
		return goja.Undefined()
	})
	_ = ab.obj.Set("raw", func(call goja.FunctionCall) goja.Value {
		addLine(call.Argument(0).String())
		return goja.Undefined()
	})

	return ab
}

func (b *sieveBuilder) String() string {
	return strings.Join(b.lines, "\n") + "\n"
}
