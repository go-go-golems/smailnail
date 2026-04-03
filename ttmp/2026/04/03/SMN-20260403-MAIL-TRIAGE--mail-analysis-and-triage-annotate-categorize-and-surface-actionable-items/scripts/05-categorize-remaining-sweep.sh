#!/usr/bin/env bash
# 05-categorize-remaining-sweep.sh — Final sweep of remaining high-volume senders
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"
COUNT=0

a() {
  local email="$1" tag="$2" note="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM annotations WHERE target_type='sender' AND target_id='$email' AND tag='$tag';")
  if [ "$existing" -gt 0 ]; then return; fi
  smailnail annotate annotation add --sqlite-path "$DB" \
    --target-type sender --target-id "$email" --tag "$tag" --note "$note" \
    --source-kind heuristic --source-label mail-triage-v1 --created-by pi-agent \
    --output json > /dev/null 2>&1
  echo "OK: $email -> $tag"
  COUNT=$((COUNT + 1))
}

echo "=== Facebook noise ==="
a "friendupdates@facebookmail.com" "noise/social-notif" "Facebook friend updates (54 msgs)"
a "reminders@facebookmail.com" "noise/social-notif" "Facebook reminders (53 msgs)"
a "memories@facebookmail.com" "noise/social-notif" "Facebook memories (40 msgs)"

echo "=== More services/commerce ==="
a "extracare@mystore.cvs.com" "noise/marketing" "CVS store marketing (53 msgs)"
a "cvs@express.medallia.com" "noise/marketing" "CVS/Medallia survey spam (45 msgs)"
a "noreply-dmarc-support@google.com" "services" "Google DMARC reports (52 msgs)"
a "no_reply@post.applecard.apple" "financial" "Apple Card statements (48 msgs)"
a "service@chewy.com" "services" "Chewy pet supplies (38 msgs)"
a "info@faphouse.com" "noise/spam" "Adult content spam (44 msgs)"

echo "=== More newsletters ==="
a "newsletter@hillelwayne.com" "newsletter/tech" "Hillel Wayne newsletter (51 msgs)"
a "andybeta@substack.com" "newsletter/culture" "Andy Beta music/culture newsletter (46 msgs)"
a "post+the-weekender@substack.com" "newsletter/culture" "The Weekender substack (40 msgs)"
a "contraptions@substack.com" "newsletter/tech" "Contraptions tech newsletter (40 msgs)"
a "simonsarris@substack.com" "newsletter/culture" "Simon Sarris essays — already done"
a "isp@substack.com" "newsletter/tech" "ISP substack"

echo "=== More hobby ==="
a "info@undergroundmusicacademy.com" "hobby" "Underground Music Academy (50 msgs)"
a "team@ecamm.com" "hobby" "Ecamm streaming tools (40 msgs)"
a "lisa@svslearn.com" "hobby" "SVS Learn art education (39 msgs)"
a "bingo@patreon.com" "services" "Patreon creator notifications (40 msgs)"

echo "=== Remaining Substack newsletters (>= 20 msgs) ==="
sqlite3 "$DB" "
SELECT sender_email, COUNT(*) as cnt 
FROM messages 
WHERE sender_domain='substack.com' AND sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
GROUP BY sender_email 
HAVING cnt >= 10
ORDER BY cnt DESC;" | while IFS='|' read -r email cnt; do
  a "$email" "newsletter/tech" "Substack newsletter ($cnt msgs)"
done

echo "=== Remaining Manning senders ==="
sqlite3 "$DB" "
SELECT sender_email FROM messages 
WHERE sender_domain='manning.com' AND sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
GROUP BY sender_email;" | while read -r email; do
  a "$email" "noise/marketing" "Manning marketing email"
done

echo "=== Remaining gmail senders (2+ msgs, likely personal) ==="
sqlite3 "$DB" "
SELECT sender_email, COUNT(*) as cnt 
FROM messages 
WHERE sender_domain='gmail.com' AND sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
GROUP BY sender_email 
HAVING cnt >= 2
ORDER BY cnt DESC;" | while IFS='|' read -r email cnt; do
  a "$email" "personal" "Gmail personal sender ($cnt msgs)"
done

echo "=== Remaining facebookmail senders ==="
sqlite3 "$DB" "
SELECT sender_email FROM messages 
WHERE sender_domain='facebookmail.com' AND sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
GROUP BY sender_email;" | while read -r email; do
  a "$email" "noise/social-notif" "Facebook notifications"
done

echo "=== Remaining Apple domains ==="
for domain in email.apple.com insideapple.apple.com post.applecard.apple; do
  sqlite3 "$DB" "
  SELECT sender_email FROM messages 
  WHERE sender_domain='$domain' AND sender_email != ''
    AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
  GROUP BY sender_email;" | while read -r email; do
    a "$email" "services" "Apple service notification"
  done
done

echo "=== Remaining ealerts.bankofamerica ==="
sqlite3 "$DB" "
SELECT sender_email FROM messages 
WHERE sender_domain='ealerts.bankofamerica.com' AND sender_email != ''
  AND NOT EXISTS (SELECT 1 FROM annotations a WHERE a.target_type='sender' AND a.target_id=messages.sender_email)
GROUP BY sender_email;" | while read -r email; do
  a "$email" "financial" "Bank of America alerts"
done

echo "=== Misc remaining ==="
a "hello@elicit.com" "newsletter/tech" "Elicit AI research assistant"
a "support@github.com" "work" "GitHub support"
a "no-reply@mail.codingcoach.io" "community" "Coding Coach mentoring platform"
a "noreply@mfa.org" "community" "Museum of Fine Arts events"
a "info@kampitakis.de" "community" "Kampitakis local community"
a "mkt@mkt.databricks.com" "noise/marketing" "Databricks marketing"
a "hello@ai.engineer" "newsletter/tech" "AI Engineer newsletter"
a "team@modal.com" "newsletter/tech" "Modal cloud compute newsletter"

echo ""
echo "=== Done. Total new: $COUNT ==="
