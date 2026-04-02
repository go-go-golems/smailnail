#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/mail.sqlite" >&2
  exit 2
fi

db="$1"
dir="$(cd "$(dirname "$0")" && pwd)"

for sql in \
  "$dir/01-schema-inventory.sql" \
  "$dir/02-message-and-thread-shape.sql" \
  "$dir/03-sender-shape.sql" \
  "$dir/04-risk-and-unsubscribe-shape.sql" \
  "$dir/05-annotation-targets.sql"
do
  echo
  echo "==> $(basename "$sql")"
  sqlite3 -header -column "$db" < "$sql"
done
