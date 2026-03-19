package smailnailmodule

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/dop251/goja"
	"github.com/go-go-golems/smailnail/pkg/mailruntime"
	"github.com/go-go-golems/smailnail/pkg/services/smailnailjs"
)

func parseCriteria(rt *goja.Runtime, value goja.Value) *smailnailjs.SearchCriteria {
	criteria := mailruntime.ParseCriteria(rt, value)
	if criteria == nil {
		return nil
	}
	ret := smailnailjs.SearchCriteria(*criteria)
	return &ret
}

func parseJSDate(value goja.Value) time.Time {
	return mailruntime.ParseJSDate(value)
}

func jsToUint32Slice(value goja.Value) []uint32 {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	exported := value.Export()
	switch v := exported.(type) {
	case []interface{}:
		ret := make([]uint32, 0, len(v))
		for _, entry := range v {
			switch n := entry.(type) {
			case int64:
				ret = append(ret, uint32(n))
			case float64:
				ret = append(ret, uint32(n))
			case int:
				ret = append(ret, uint32(n))
			case uint32:
				ret = append(ret, n)
			}
		}
		return ret
	case int64:
		return []uint32{uint32(v)}
	case float64:
		return []uint32{uint32(v)}
	default:
		if obj, ok := value.(*goja.Object); ok && obj.Get("length") != nil {
			length := int(obj.Get("length").ToInteger())
			ret := make([]uint32, 0, length)
			for i := 0; i < length; i++ {
				ret = append(ret, uint32(obj.Get(strconv.Itoa(i)).ToInteger()))
			}
			return ret
		}
		return []uint32{uint32(value.ToInteger())}
	}
}

func jsToFetchFields(value goja.Value) []smailnailjs.FetchField {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return []smailnailjs.FetchField{
			smailnailjs.FetchUID,
			smailnailjs.FetchFlags,
			smailnailjs.FetchEnvelope,
			smailnailjs.FetchInternalDate,
			smailnailjs.FetchSize,
		}
	}
	strs := jsToStringSlice(value)
	ret := make([]smailnailjs.FetchField, 0, len(strs))
	for _, s := range strs {
		ret = append(ret, smailnailjs.FetchField(s))
	}
	return ret
}

func jsToStringSlice(value goja.Value) []string {
	if goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	exported := value.Export()
	switch v := exported.(type) {
	case []interface{}:
		ret := make([]string, len(v))
		for i, entry := range v {
			ret[i] = fmt.Sprintf("%v", entry)
		}
		return ret
	case string:
		return []string{v}
	default:
		if obj, ok := value.(*goja.Object); ok && obj.Get("length") != nil {
			length := int(obj.Get("length").ToInteger())
			ret := make([]string, 0, length)
			for i := 0; i < length; i++ {
				ret = append(ret, obj.Get(strconv.Itoa(i)).String())
			}
			return ret
		}
		return []string{fmt.Sprintf("%v", exported)}
	}
}

func readSilentOption(rt *goja.Runtime, call goja.FunctionCall, argIndex int) bool {
	silent := false
	if len(call.Arguments) > argIndex && !goja.IsUndefined(call.Arguments[argIndex]) && !goja.IsNull(call.Arguments[argIndex]) {
		options := call.Arguments[argIndex].ToObject(rt)
		if v := options.Get("silent"); v != nil && !goja.IsUndefined(v) {
			silent = v.ToBoolean()
		}
	}
	return silent
}

func toJSONValue(rt *goja.Runtime, value interface{}) goja.Value {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(rt.NewGoError(err))
	}
	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		panic(rt.NewGoError(err))
	}
	return rt.ToValue(decoded)
}
