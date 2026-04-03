#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="${BASE_DIR:-/tmp/smailnail-last-6-months-parallel}"
SESSION_PREFIX="${SESSION_PREFIX:-smn-mirror-month}"

echo "tmux sessions:"
tmux list-sessions 2>/dev/null | rg "${SESSION_PREFIX}-" || true
echo

shopt -s nullglob
for log_path in "${BASE_DIR}"/*.log; do
  shard="$(basename "${log_path%.log}")"
  sqlite_path="${BASE_DIR}/${shard}.sqlite"
  out_path="${BASE_DIR}/${shard}.out.json"

  echo "${shard}:"
  if [[ -f "$sqlite_path" ]]; then
    echo "  rows: $(sqlite3 "$sqlite_path" 'select count(*) from messages;' 2>/dev/null || echo '?')"
  else
    echo "  rows: not created yet"
  fi

  if [[ -f "$out_path" ]]; then
    echo "  final output:"
    tail -n 20 "$out_path" | sed 's/^/    /'
  else
    echo "  final output: pending"
  fi

  echo "  recent log:"
  tail -n 10 "$log_path" | sed 's/^/    /'
  echo
done
