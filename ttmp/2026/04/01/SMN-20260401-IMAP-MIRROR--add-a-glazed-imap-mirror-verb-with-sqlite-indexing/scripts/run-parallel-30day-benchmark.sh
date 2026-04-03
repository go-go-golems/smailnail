#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail"

SESSION_A="smn-mirror-a"
SESSION_B="smn-mirror-b"

SQLITE_A="/tmp/smailnail-parallel-a.sqlite"
SQLITE_B="/tmp/smailnail-parallel-b.sqlite"
RAW_A="/tmp/smailnail-parallel-a-raw"
RAW_B="/tmp/smailnail-parallel-b-raw"

OUT_A="/tmp/smailnail-parallel-a.out"
OUT_B="/tmp/smailnail-parallel-b.out"
LOG_A="/tmp/smailnail-parallel-a.log"
LOG_B="/tmp/smailnail-parallel-b.log"

rm -rf "$SQLITE_A" "$SQLITE_B" "$RAW_A" "$RAW_B" "$OUT_A" "$OUT_B" "$LOG_A" "$LOG_B"

tmux kill-session -t "$SESSION_A" >/dev/null 2>&1 || true
tmux kill-session -t "$SESSION_B" >/dev/null 2>&1 || true

tmux new-session -d -s "$SESSION_A" \
  "cd '$REPO_ROOT' && set -a && source .envrc && set +a && /usr/bin/time -p go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror --server mail.bl0rg.net --username manuel --password \"\$MAIL_PASSWORD\" --mailbox INBOX --since-days 30 --sqlite-path '$SQLITE_A' --mirror-root '$RAW_A' --output json > '$OUT_A' 2> '$LOG_A'"

tmux new-session -d -s "$SESSION_B" \
  "cd '$REPO_ROOT' && set -a && source .envrc && set +a && /usr/bin/time -p go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror --server mail.bl0rg.net --username manuel --password \"\$MAIL_PASSWORD\" --mailbox INBOX --since-days 30 --sqlite-path '$SQLITE_B' --mirror-root '$RAW_B' --output json > '$OUT_B' 2> '$LOG_B'"

echo "Started $SESSION_A and $SESSION_B"
echo "Monitor with:"
echo "  $(dirname "$0")/check-parallel-30day-benchmark.sh"
