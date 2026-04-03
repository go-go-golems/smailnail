#!/usr/bin/env bash
set -euo pipefail

SESSION_A="smn-mirror-a"
SESSION_B="smn-mirror-b"

SQLITE_A="/tmp/smailnail-parallel-a.sqlite"
SQLITE_B="/tmp/smailnail-parallel-b.sqlite"
OUT_A="/tmp/smailnail-parallel-a.out"
OUT_B="/tmp/smailnail-parallel-b.out"
LOG_A="/tmp/smailnail-parallel-a.log"
LOG_B="/tmp/smailnail-parallel-b.log"

echo "tmux sessions:"
tmux list-sessions 2>/dev/null | rg 'smn-mirror-(a|b)' || true
echo

for label in a b; do
  upper_label=$(printf '%s' "$label" | tr '[:lower:]' '[:upper:]')
  sqlite_var="SQLITE_${upper_label}"
  out_var="OUT_${upper_label}"
  log_var="LOG_${upper_label}"
  sqlite_path="${!sqlite_var}"
  out_path="${!out_var}"
  log_path="${!log_var}"

  echo "mirror-$label:"
  if [[ -f "$sqlite_path" ]]; then
    echo "  sqlite: $sqlite_path"
    echo "  rows: $(sqlite3 "$sqlite_path" "select count(*) from messages;" 2>/dev/null || echo '?')"
  else
    echo "  sqlite: not created yet"
  fi

  if [[ -f "$out_path" ]]; then
    echo "  final output:"
    tail -n 20 "$out_path" | sed 's/^/    /'
  else
    echo "  final output: pending"
  fi

  if [[ -f "$log_path" ]]; then
    echo "  recent log:"
    tail -n 10 "$log_path" | sed 's/^/    /'
  else
    echo "  recent log: pending"
  fi
  echo
done
