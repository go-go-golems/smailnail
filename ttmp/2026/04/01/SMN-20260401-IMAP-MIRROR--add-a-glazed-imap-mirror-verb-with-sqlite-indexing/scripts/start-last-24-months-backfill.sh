#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANAGER_SESSION="${MANAGER_SESSION:-smn-backfill-24m-manager}"
DASHBOARD_SESSION="${DASHBOARD_SESSION:-smn-backfill-24m-dashboard}"

if tmux has-session -t "$MANAGER_SESSION" 2>/dev/null; then
  echo "tmux session $MANAGER_SESSION already exists" >&2
  exit 1
fi

if tmux has-session -t "$DASHBOARD_SESSION" 2>/dev/null; then
  echo "tmux session $DASHBOARD_SESSION already exists" >&2
  exit 1
fi

tmux new-session -d -s "$MANAGER_SESSION" "${SCRIPT_DIR}/run-last-24-months-backfill.sh"
tmux new-session -d -s "$DASHBOARD_SESSION" "${SCRIPT_DIR}/dashboard-last-24-months-backfill.sh"

echo "started:"
echo "  manager:   ${MANAGER_SESSION}"
echo "  dashboard: ${DASHBOARD_SESSION}"
echo
echo "attach with:"
echo "  tmux attach -t ${DASHBOARD_SESSION}"
