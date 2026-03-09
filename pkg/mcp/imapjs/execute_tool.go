package imapjs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
	ggjengine "github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	smailnailmodule "github.com/go-go-golems/smailnail/pkg/js/modules/smailnail"
)

func executeIMAPJSHandler(ctx context.Context, raw map[string]interface{}) (*protocol.ToolResult, error) {
	args := embeddable.NewArguments(raw)

	var req ExecuteIMAPJSRequest
	if err := args.BindArguments(&req); err != nil {
		return newErrorToolResult("invalid arguments", err), nil
	}
	if req.Code == "" {
		return newErrorToolResult("code is required", nil), nil
	}

	module := smailnailmodule.NewModule()
	factory, err := ggjengine.NewBuilder().
		WithModules(ggjengine.NativeModuleSpec{
			ModuleName: module.Name(),
			Loader:     module.Loader,
		}).
		Build()
	if err != nil {
		return newErrorToolResult("failed to build runtime", err), nil
	}

	rt, err := factory.NewRuntime(ctx)
	if err != nil {
		return newErrorToolResult("failed to create runtime", err), nil
	}
	defer func() {
		_ = rt.Close(context.Background())
	}()

	value, err := rt.VM.RunString(req.Code)
	if err != nil {
		return newErrorToolResult("JavaScript execution failed", err), nil
	}

	response := ExecuteIMAPJSResponse{
		Success: true,
		Value:   exportValue(value),
		Console: []string{},
	}
	return newJSONToolResult(response)
}

func exportValue(value goja.Value) any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	return value.Export()
}

func newJSONToolResult(v any) (*protocol.ToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal tool response: %w", err)
	}
	return protocol.NewToolResult(protocol.WithText(string(data))), nil
}

func newErrorToolResult(message string, err error) *protocol.ToolResult {
	toolErr := &ToolError{Message: message}
	if err != nil {
		toolErr.Kind = fmt.Sprintf("%T", err)
		toolErr.Message = fmt.Sprintf("%s: %v", message, err)
	}
	res, marshalErr := newJSONToolResult(ExecuteIMAPJSResponse{
		Success: false,
		Error:   toolErr,
		Console: []string{},
	})
	if marshalErr != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(toolErr.Message))
	}
	res.IsError = true
	return res
}
