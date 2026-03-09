package imapjs

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func getIMAPJSDocumentationHandler(_ context.Context, raw map[string]interface{}) (*protocol.ToolResult, error) {
	args := embeddable.NewArguments(raw)

	var req DocumentationRequest
	if err := args.BindArguments(&req); err != nil {
		return newErrorToolResult("invalid documentation query arguments", err), nil
	}

	registry, err := getDefaultDocsRegistry()
	if err != nil {
		return newErrorToolResult("failed to load embedded documentation", err), nil
	}

	response, err := registry.query(req)
	if err != nil {
		return newErrorToolResult("documentation query failed", err), nil
	}
	return newJSONToolResult(response)
}
