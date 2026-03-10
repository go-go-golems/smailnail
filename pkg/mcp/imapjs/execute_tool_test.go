package imapjs

import (
	"context"
	"encoding/json"
	"testing"
)

func TestExecuteIMAPJSHandlerBuildRule(t *testing.T) {
	result, err := executeIMAPJSHandler(context.Background(), map[string]interface{}{
		"code": `
const smailnail = require("smailnail");
const rule = smailnail.buildRule({
  name: "invoice-search",
  subjectContains: "invoice",
  includeContent: true,
  contentType: "text/plain"
});
({
  subjectContains: rule.search.subjectContains,
  contentType: rule.output.fields[7].content.types[0]
});
`,
	})
	if err != nil {
		t.Fatalf("executeIMAPJSHandler returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got error: %#v", result)
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected one content item, got %d", len(result.Content))
	}

	var decoded ExecuteIMAPJSResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !decoded.Success {
		t.Fatalf("expected success response, got %#v", decoded)
	}

	value, ok := decoded.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("value type = %T, want map[string]interface{}", decoded.Value)
	}
	if got := value["subjectContains"]; got != "invoice" {
		t.Fatalf("subjectContains = %#v, want %q", got, "invoice")
	}
	if got := value["contentType"]; got != "text/plain" {
		t.Fatalf("contentType = %#v, want %q", got, "text/plain")
	}
}

func TestExecuteIMAPJSHandlerReturnsStructuredError(t *testing.T) {
	result, err := executeIMAPJSHandler(context.Background(), map[string]interface{}{
		"code": `throw new Error("boom");`,
	})
	if err != nil {
		t.Fatalf("executeIMAPJSHandler returned error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error result")
	}

	var decoded ExecuteIMAPJSResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decoded.Success {
		t.Fatalf("expected unsuccessful response")
	}
	if decoded.Error == nil {
		t.Fatalf("expected error payload")
	}
}
