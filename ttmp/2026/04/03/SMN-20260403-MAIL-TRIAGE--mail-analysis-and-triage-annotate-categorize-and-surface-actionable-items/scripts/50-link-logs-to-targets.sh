#!/usr/bin/env bash
# 50-link-logs-to-targets.sh — Link annotation log entries to the targets they describe
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"

# Log 1: "Mail triage research: mailbox profile and annotation plan"
# This was the initial research — link to the overall account
LOG1="428937da-c651-497a-9985-04a9a77352ce"

# Log 2: "Mail triage v1 complete: 416 annotations, 5 groups, summary report generated"
# Link to all annotation categories (by domain-level targets representing the work)
LOG2="09377b5f-9bcf-4be0-8de7-5c3e47971c8e"

# Log 3: "Added high-importance tags: tax, legal, equity, housing, health, conferences"
LOG3="44b5a9dc-6224-41f4-901f-3d87b573891d"

# Log 4: "Embedding/RAG design"
LOG4="a7296f62-521a-4e1a-884f-0af3cf1e58ed"

link() {
  local log_id="$1" target_type="$2" target_id="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotation_log_targets WHERE log_id='$log_id' AND target_type='$target_type' AND target_id='$target_id';")
  if [ "$existing" -gt 0 ]; then return; fi
  smailnail annotate log link-target --sqlite-path "$DB" \
    --log-id "$log_id" --target-type "$target_type" --target-id "$target_id" \
    --output json > /dev/null 2>&1
  echo "OK: log $log_id -> $target_type:$target_id"
}

echo "=== Log 1: Research phase ==="
link "$LOG1" "account" "manuel"

echo ""
echo "=== Log 2: Triage completion — link to key noise senders ==="
link "$LOG2" "sender" "notifications@github.com"
link "$LOG2" "sender" "sales@thetreecenter.com"
link "$LOG2" "sender" "hello@readwise.io"
link "$LOG2" "sender" "ginny.white@gmail.com"
link "$LOG2" "sender" "notifier@mail.rollbar.com"
link "$LOG2" "sender" "intern@lists.entropia.de"

echo ""
echo "=== Log 3: High-importance senders ==="
link "$LOG3" "sender" "dave@davemillercpa.com"
link "$LOG3" "sender" "liam@davemillercpa.com"
link "$LOG3" "sender" "office@davemillercpa.com"
link "$LOG3" "sender" "kryzaklaw@yahoo.com"
link "$LOG3" "sender" "ellen@wrightfamilylawgroup.com"
link "$LOG3" "sender" "elaine@wrightfamilylawgroup.com"
link "$LOG3" "sender" "tripp@equitybee.com"
link "$LOG3" "sender" "supply@equityzen.com"
link "$LOG3" "sender" "noreply@mail.hiive.com"
link "$LOG3" "sender" "donotreply@admoove.com"
link "$LOG3" "sender" "no-reply@carta.com"
link "$LOG3" "sender" "notifications@alerts.biltrewards.com"
link "$LOG3" "sender" "halsteadprovidence@bozzuto.com"
link "$LOG3" "sender" "noreply@prosperhealth.io"
link "$LOG3" "sender" "demetrios@news.mlops.community"
link "$LOG3" "sender" "info@ai.engineer"

echo ""
echo "=== Log 4: Embedding design — link to account (affects whole mailbox) ==="
link "$LOG4" "account" "manuel"

echo ""
echo "=== Verify ==="
sqlite3 -header -column "$DB" "
SELECT l.title, lt.target_type, lt.target_id
FROM annotation_log_targets lt
JOIN annotation_logs l ON l.id = lt.log_id
ORDER BY l.created_at, lt.target_type, lt.target_id;"
