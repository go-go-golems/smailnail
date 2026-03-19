package imapjs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoadDocsRegistry(t *testing.T) {
	registry, err := loadDocsRegistry()
	if err != nil {
		t.Fatalf("loadDocsRegistry returned error: %v", err)
	}
	if registry.store.ByPackage["smailnail"] == nil {
		t.Fatalf("expected smailnail package doc")
	}
	if registry.store.BySymbol["connect"] == nil {
		t.Fatalf("expected connect symbol doc")
	}
	if registry.store.ByExample["imap-session-automation"] == nil {
		t.Fatalf("expected imap-session-automation example")
	}
}

func TestGetIMAPJSDocumentationSymbolMode(t *testing.T) {
	result, err := getIMAPJSDocumentationHandler(context.Background(), map[string]interface{}{
		"mode":        "symbol",
		"symbol":      "connect",
		"includeBody": true,
	})
	if err != nil {
		t.Fatalf("getIMAPJSDocumentationHandler returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got error")
	}

	var decoded DocumentationResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decoded.Mode != "symbol" {
		t.Fatalf("mode = %q, want %q", decoded.Mode, "symbol")
	}
	if len(decoded.Symbols) != 1 || decoded.Symbols[0].Name != "connect" {
		t.Fatalf("unexpected symbols: %#v", decoded.Symbols)
	}
	found := false
	for _, example := range decoded.Examples {
		if example.ID == "imap-session-automation" {
			found = true
			if example.Body == "" {
				t.Fatalf("expected example body to be included")
			}
		}
	}
	if !found {
		t.Fatalf("expected imap-session-automation example, got %#v", decoded.Examples)
	}
}

func TestGetIMAPJSDocumentationRenderMode(t *testing.T) {
	result, err := getIMAPJSDocumentationHandler(context.Background(), map[string]interface{}{
		"mode": "render",
	})
	if err != nil {
		t.Fatalf("getIMAPJSDocumentationHandler returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success result, got error")
	}

	var decoded DocumentationResponse
	if err := json.Unmarshal([]byte(result.Content[0].Text), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !strings.Contains(decoded.RenderedMarkdown, "Package: smailnail") {
		t.Fatalf("expected rendered markdown to contain package heading")
	}
	if !strings.Contains(decoded.RenderedMarkdown, "Symbol: connect") {
		t.Fatalf("expected rendered markdown to contain symbol heading")
	}
}
