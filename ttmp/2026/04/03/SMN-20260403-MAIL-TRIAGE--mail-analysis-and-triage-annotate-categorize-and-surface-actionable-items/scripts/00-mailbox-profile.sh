#!/usr/bin/env bash
# 00-mailbox-profile.sh — Generate a mailbox profile summary
# Run from the smailnail project root
set -euo pipefail

DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"

echo "=== Mailbox Profile ==="
echo ""
echo "--- Total messages ---"
sqlite3 "$DB" "SELECT COUNT(*) FROM messages;"
echo ""
echo "--- Date range ---"
sqlite3 "$DB" "SELECT MIN(internal_date) || ' to ' || MAX(internal_date) FROM messages;"
echo ""
echo "--- Messages per month ---"
sqlite3 -header -column "$DB" "
SELECT strftime('%Y-%m', internal_date) as month, COUNT(*) as cnt
FROM messages GROUP BY month ORDER BY month;"
echo ""
echo "--- Top 30 sender domains ---"
sqlite3 -header -column "$DB" "
SELECT sender_domain, COUNT(*) as cnt
FROM messages WHERE sender_domain != ''
GROUP BY sender_domain ORDER BY cnt DESC LIMIT 30;"
echo ""
echo "--- Top 30 sender emails ---"
sqlite3 -header -column "$DB" "
SELECT sender_email, COUNT(*) as cnt
FROM messages WHERE sender_email != ''
GROUP BY sender_email ORDER BY cnt DESC LIMIT 30;"
echo ""
echo "--- Thread size distribution ---"
sqlite3 -header -column "$DB" "
SELECT
  CASE
    WHEN message_count = 1 THEN '1 msg'
    WHEN message_count BETWEEN 2 AND 5 THEN '2-5 msgs'
    WHEN message_count BETWEEN 6 AND 20 THEN '6-20 msgs'
    WHEN message_count > 20 THEN '20+ msgs'
  END as thread_size,
  COUNT(*) as num_threads
FROM threads GROUP BY thread_size ORDER BY num_threads DESC;"
echo ""
echo "--- Enrichment status ---"
echo "Threads: $(sqlite3 "$DB" "SELECT COUNT(*) FROM threads;")"
echo "Senders: $(sqlite3 "$DB" "SELECT COUNT(*) FROM senders;")"
echo "Annotations: $(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations;")"
echo "Groups: $(sqlite3 "$DB" "SELECT COUNT(*) FROM target_groups;")"
echo "Logs: $(sqlite3 "$DB" "SELECT COUNT(*) FROM annotation_logs;")"
echo ""
echo "--- Senders with unsubscribe ---"
sqlite3 "$DB" "SELECT COUNT(*) FROM senders WHERE has_list_unsubscribe = 1;"
