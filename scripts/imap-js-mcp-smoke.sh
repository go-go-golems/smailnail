#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail"

cd "$ROOT"

go test ./pkg/mcp/imapjs ./pkg/js/modules/smailnail -count=1
go build ./cmd/smailnail-imap-mcp

tool_listing="$(go run ./cmd/smailnail-imap-mcp mcp list-tools)"
printf '%s\n' "$tool_listing"
printf '%s\n' "$tool_listing" | grep -F 'Tool: executeIMAPJS' >/dev/null
printf '%s\n' "$tool_listing" | grep -F 'Tool: getIMAPJSDocumentation' >/dev/null

echo "smailnail IMAP JS MCP smoke passed."
