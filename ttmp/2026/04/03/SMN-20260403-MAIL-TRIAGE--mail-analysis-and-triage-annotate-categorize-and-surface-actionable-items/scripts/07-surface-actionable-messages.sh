#!/usr/bin/env bash
# 07-surface-actionable-messages.sh — Surface interesting/actionable messages and threads
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"
OUT_DIR="$(dirname "$0")/../various"
mkdir -p "$OUT_DIR"

echo "=== Generating message surface reports ==="

echo "--- 1. Recent personal emails (last 90 days) that may need a reply ---"
sqlite3 -header -column "$DB" "
SELECT m.id, substr(m.internal_date,1,10) as date, m.sender_email, 
       substr(m.subject,1,70) as subject
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='personal'
WHERE m.internal_date >= date('now', '-90 days')
ORDER BY m.internal_date DESC;" | tee "$OUT_DIR/recent-personal.txt"

echo ""
echo "--- 2. Multi-message personal threads (real conversations) ---"
sqlite3 -header -column "$DB" "
SELECT t.thread_id, t.subject, t.message_count, t.participant_count,
       substr(t.first_sent_date,1,10) as first, substr(t.last_sent_date,1,10) as last
FROM threads t
WHERE t.message_count >= 3
  AND EXISTS (
    SELECT 1 FROM messages m 
    JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='personal'
    WHERE m.thread_id = t.thread_id
  )
ORDER BY t.last_sent_date DESC;" | tee "$OUT_DIR/personal-threads.txt"

echo ""
echo "--- 3. Work threads from last 90 days ---"
sqlite3 -header -column "$DB" "
SELECT m.id, substr(m.internal_date,1,10) as date, m.sender_email, 
       substr(m.subject,1,70) as subject
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='work'
WHERE m.internal_date >= date('now', '-90 days')
  AND m.sender_domain NOT IN ('github.com')
ORDER BY m.internal_date DESC
LIMIT 30;" | tee "$OUT_DIR/recent-work.txt"

echo ""
echo "--- 4. Community messages from last 60 days ---"
sqlite3 -header -column "$DB" "
SELECT m.id, substr(m.internal_date,1,10) as date, m.sender_email, 
       substr(m.subject,1,70) as subject
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='community'
WHERE m.internal_date >= date('now', '-60 days')
ORDER BY m.internal_date DESC
LIMIT 30;" | tee "$OUT_DIR/recent-community.txt"

echo ""
echo "--- 5. Top newsletter issues from last 30 days ---"
sqlite3 -header -column "$DB" "
SELECT m.id, substr(m.internal_date,1,10) as date, m.sender_email, 
       substr(m.subject,1,70) as subject
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='newsletter/tech'
WHERE m.internal_date >= date('now', '-30 days')
ORDER BY m.internal_date DESC
LIMIT 40;" | tee "$OUT_DIR/recent-tech-newsletters.txt"

echo ""
echo "--- 6. Financial alerts from last 60 days ---"
sqlite3 -header -column "$DB" "
SELECT m.id, substr(m.internal_date,1,10) as date, m.sender_email, 
       substr(m.subject,1,70) as subject
FROM messages m
JOIN annotations a ON a.target_type='sender' AND a.target_id=m.sender_email AND a.tag='financial'
WHERE m.internal_date >= date('now', '-60 days')
ORDER BY m.internal_date DESC
LIMIT 20;" | tee "$OUT_DIR/recent-financial.txt"

echo ""
echo "=== Reports saved to $OUT_DIR ==="
