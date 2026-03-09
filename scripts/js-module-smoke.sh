#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail"

cd "$ROOT"

go test ./pkg/services/smailnailjs ./pkg/js/modules/smailnail -count=1

echo "smailnail JS module smoke passed."
