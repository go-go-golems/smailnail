#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT"

go test ./pkg/mcp/imapjs ./pkg/js/modules/smailnail -count=1
go build ./cmd/smailnail-imap-mcp

tool_listing="$(go run ./cmd/smailnail-imap-mcp mcp list-tools)"
printf '%s\n' "$tool_listing"
printf '%s\n' "$tool_listing" | grep -F 'Tool: executeIMAPJS' >/dev/null
printf '%s\n' "$tool_listing" | grep -F 'Tool: getIMAPJSDocumentation' >/dev/null

echo "smailnail IMAP JS MCP smoke passed."
