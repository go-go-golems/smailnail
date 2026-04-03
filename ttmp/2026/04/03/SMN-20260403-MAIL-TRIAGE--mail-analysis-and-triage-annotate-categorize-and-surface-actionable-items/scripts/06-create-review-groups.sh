#!/usr/bin/env bash
# 06-create-review-groups.sh — Create target groups for review clusters
set -euo pipefail
DB="${SMAILNAIL_DB:-$HOME/smailnail/smailnail-last-24-months-merged.sqlite}"

create_group() {
  local name="$1" desc="$2"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM target_groups WHERE name='$name';")
  if [ "$existing" -gt 0 ]; then
    echo "SKIP group: $name (exists)"
    sqlite3 "$DB" "SELECT id FROM target_groups WHERE name='$name';"
    return
  fi
  local id
  id=$(smailnail annotate group create --sqlite-path "$DB" \
    --name "$name" --description "$desc" \
    --source-kind agent --source-label mail-triage-v1 --created-by pi-agent \
    --select id 2>/dev/null)
  echo "OK group: $name -> $id"
  echo "$id"
}

add_target() {
  local gid="$1" ttype="$2" tid="$3"
  existing=$(sqlite3 "$DB" "SELECT COUNT(*) FROM target_group_members WHERE group_id='$gid' AND target_type='$ttype' AND target_id='$tid';")
  if [ "$existing" -gt 0 ]; then return; fi
  smailnail annotate group add-target --sqlite-path "$DB" \
    --group-id "$gid" --target-type "$ttype" --target-id "$tid" \
    --output json > /dev/null 2>&1
}

echo "=== Creating groups ==="

# 1. Unsubscribe candidates
UNSUB_GID=$(create_group "Unsubscribe Candidates" "High-volume senders with List-Unsubscribe headers — candidates for unsubscribing")
# Add noise senders that have unsubscribe
sqlite3 "$DB" "
SELECT a.target_id FROM annotations a
JOIN senders s ON s.email = a.target_id
WHERE a.target_type = 'sender'
  AND a.tag LIKE 'noise/%'
  AND s.has_list_unsubscribe = 1
ORDER BY s.msg_count DESC;" | while read -r email; do
  add_target "$UNSUB_GID" "sender" "$email"
  echo "  + unsub candidate: $email"
done

echo ""
# 2. Valuable newsletters
NL_GID=$(create_group "Valuable Newsletters" "Newsletters tagged as worth reading — tech, culture, creative")
sqlite3 "$DB" "
SELECT target_id FROM annotations
WHERE target_type = 'sender' AND tag LIKE 'newsletter/%'
ORDER BY target_id;" | while read -r email; do
  add_target "$NL_GID" "sender" "$email"
  echo "  + newsletter: $email"
done

echo ""
# 3. Personal contacts
PC_GID=$(create_group "Personal Contacts" "Real human correspondents")
sqlite3 "$DB" "
SELECT target_id FROM annotations
WHERE target_type = 'sender' AND tag = 'personal'
ORDER BY target_id;" | while read -r email; do
  add_target "$PC_GID" "sender" "$email"
  echo "  + personal: $email"
done

echo ""
# 4. Active work threads
WORK_GID=$(create_group "Work Senders" "Work-related senders: team-mento, rollbar, slack, zulip")
sqlite3 "$DB" "
SELECT target_id FROM annotations
WHERE target_type = 'sender' AND tag = 'work'
ORDER BY target_id;" | while read -r email; do
  add_target "$WORK_GID" "sender" "$email"
  echo "  + work: $email"
done

echo ""
# 5. Hobby & creative
HOBBY_GID=$(create_group "Hobby & Creative" "Music, photography, 3D printing, gaming, fitness, art")
sqlite3 "$DB" "
SELECT target_id FROM annotations
WHERE target_type = 'sender' AND tag IN ('hobby', 'newsletter/creative')
ORDER BY target_id;" | while read -r email; do
  add_target "$HOBBY_GID" "sender" "$email"
  echo "  + hobby: $email"
done

echo ""
echo "=== Groups created ==="
smailnail annotate group list --sqlite-path "$DB" --output json 2>/dev/null
