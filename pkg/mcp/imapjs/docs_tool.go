package imapjs

import (
	"context"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func getIMAPJSDocumentationHandler(_ context.Context, _ map[string]interface{}) (*protocol.ToolResult, error) {
	return newJSONToolResult(map[string]any{
		"implemented": false,
		"message":     "getIMAPJSDocumentation is not implemented yet in this slice",
	})
}
