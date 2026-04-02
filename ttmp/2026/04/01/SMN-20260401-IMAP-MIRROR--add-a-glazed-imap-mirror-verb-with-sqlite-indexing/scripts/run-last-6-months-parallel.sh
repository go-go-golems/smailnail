#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail}"
BASE_DIR="${BASE_DIR:-/tmp/smailnail-last-6-months-parallel}"
SERVER="${SERVER:-mail.bl0rg.net}"
USERNAME="${USERNAME:-manuel}"
MAILBOX="${MAILBOX:-INBOX}"
MONTHS="${MONTHS:-6}"
SESSION_PREFIX="${SESSION_PREFIX:-smn-mirror-month}"

mkdir -p "$BASE_DIR"

month_zero="$(date -u +%Y-%m-01)"

for ((offset=MONTHS-1; offset>=0; offset--)); do
  start="$(date -u -d "$month_zero -$offset month" +%F)"
  end="$(date -u -d "$start +1 month" +%F)"
  shard="$(date -u -d "$start" +%Y-%m)"
  session="${SESSION_PREFIX}-${shard}"
  sqlite_path="${BASE_DIR}/${shard}.sqlite"
  mirror_root="${BASE_DIR}/${shard}-raw"
  out_path="${BASE_DIR}/${shard}.out.json"
  log_path="${BASE_DIR}/${shard}.log"

  rm -rf "$sqlite_path" "$mirror_root" "$out_path" "$log_path"
  tmux kill-session -t "$session" >/dev/null 2>&1 || true

  printf -v cmd 'cd %q && set -a && source .envrc && set +a && /usr/bin/time -p go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror --server %q --username %q --password "$MAIL_PASSWORD" --mailbox %q --since-date %q --before-date %q --sqlite-path %q --mirror-root %q --output json > %q 2> %q' \
    "$REPO_ROOT" "$SERVER" "$USERNAME" "$MAILBOX" "$start" "$end" "$sqlite_path" "$mirror_root" "$out_path" "$log_path"

  tmux new-session -d -s "$session" "$cmd"
  echo "started ${session}: ${start} <= date < ${end}"
done

echo
echo "Monitor with:"
echo "  $(dirname "$0")/check-last-6-months-parallel.sh"
echo
echo "Artifacts:"
echo "  base dir: $BASE_DIR"
