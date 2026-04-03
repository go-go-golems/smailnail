#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="${BASE_DIR:-/tmp/smailnail-last-24-months-backfill}"
REFRESH_SECONDS="${REFRESH_SECONDS:-5}"

format_elapsed() {
  python - "$1" "$2" <<'PY'
import datetime as dt
import sys

start_text = sys.argv[1]
end_text = sys.argv[2]
if not start_text:
    print("-")
    raise SystemExit

def parse(value):
    return dt.datetime.fromisoformat(value.replace("Z", "+00:00"))

start = parse(start_text)
end = parse(end_text) if end_text else dt.datetime.now(dt.timezone.utc)
seconds = int((end - start).total_seconds())
h, rem = divmod(seconds, 3600)
m, s = divmod(rem, 60)
if h:
    print(f"{h:02d}:{m:02d}:{s:02d}")
else:
    print(f"{m:02d}:{s:02d}")
PY
}

while true; do
  clear
  echo "smailnail 24-month backfill dashboard"
  echo "updated: $(date)"
  echo "base dir: ${BASE_DIR}"
  echo
  printf "%-8s %-10s %-7s %-8s %-10s %s\n" "shard" "state" "rows" "elapsed" "exit" "recent"
  printf "%-8s %-10s %-7s %-8s %-10s %s\n" "--------" "----------" "-------" "--------" "----------" "------"

  if [[ -f "${BASE_DIR}/manifest.tsv" ]]; then
    while IFS=$'\t' read -r shard start end; do
      shard_dir="${BASE_DIR}/${shard}"
      state="pending"
      rows="-"
      elapsed="-"
      exit_code="-"
      recent="-"

      if [[ -f "${shard_dir}/state" ]]; then
        state="$(<"${shard_dir}/state")"
      fi
      if [[ -f "${shard_dir}/mirror.sqlite" ]]; then
        rows="$(sqlite3 "${shard_dir}/mirror.sqlite" 'select count(*) from messages;' 2>/dev/null || echo '?')"
      fi
      if [[ -f "${shard_dir}/started_at" ]]; then
        started_at="$(<"${shard_dir}/started_at")"
        finished_at=""
        if [[ -f "${shard_dir}/finished_at" ]]; then
          finished_at="$(<"${shard_dir}/finished_at")"
        fi
        elapsed="$(format_elapsed "$started_at" "$finished_at")"
      fi
      if [[ -f "${shard_dir}/exit_code" ]]; then
        exit_code="$(<"${shard_dir}/exit_code")"
      fi
      if [[ -f "${shard_dir}/run.log" ]]; then
        recent="$(tail -n 1 "${shard_dir}/run.log" | cut -c1-90)"
      fi

      printf "%-8s %-10s %-7s %-8s %-10s %s\n" "$shard" "$state" "$rows" "$elapsed" "$exit_code" "$recent"
    done <"${BASE_DIR}/manifest.tsv"
  else
    echo "manifest not written yet"
  fi

  echo
  if [[ -f "${BASE_DIR}/manager.log" ]]; then
    echo "manager:"
    tail -n 5 "${BASE_DIR}/manager.log"
  fi

  sleep "$REFRESH_SECONDS"
done
