#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="${REPO_ROOT:-/home/manuel/workspaces/2026-04-01/smailnail-sqlite/smailnail}"
BASE_DIR="${BASE_DIR:-/tmp/smailnail-last-24-months-backfill}"
SERVER="${SERVER:-mail.bl0rg.net}"
USERNAME="${USERNAME:-manuel}"
MAILBOX="${MAILBOX:-INBOX}"
MONTHS="${MONTHS:-24}"
PARALLEL="${PARALLEL:-6}"

mkdir -p "$BASE_DIR"

if [[ ! -f "${REPO_ROOT}/.envrc" ]]; then
  echo "missing ${REPO_ROOT}/.envrc" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "${REPO_ROOT}/.envrc"
set +a

if [[ -z "${MAIL_PASSWORD:-}" ]]; then
  echo "MAIL_PASSWORD is not set after sourcing .envrc" >&2
  exit 1
fi

manifest="${BASE_DIR}/manifest.tsv"
manager_log="${BASE_DIR}/manager.log"
: >"$manifest"
: >"$manager_log"

month_zero="$(date -u +%Y-%m-01)"

log_manager() {
  printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$manager_log"
}

format_month_start() {
  local offset="$1"
  date -u -d "${month_zero} -${offset} month" +%F
}

write_shard_manifest() {
  local start end shard
  for ((offset=MONTHS-1; offset>=0; offset--)); do
    start="$(format_month_start "$offset")"
    end="$(date -u -d "${start} +1 month" +%F)"
    shard="$(date -u -d "$start" +%Y-%m)"
    printf '%s\t%s\t%s\n' "$shard" "$start" "$end" >>"$manifest"
  done
}

run_shard() {
  local shard="$1"
  local start="$2"
  local end="$3"
  local shard_dir="${BASE_DIR}/${shard}"
  local sqlite_path="${shard_dir}/mirror.sqlite"
  local mirror_root="${shard_dir}/raw"
  local output_path="${shard_dir}/result.json"
  local log_path="${shard_dir}/run.log"
  local state_path="${shard_dir}/state"
  local exit_path="${shard_dir}/exit_code"
  local started_path="${shard_dir}/started_at"
  local finished_path="${shard_dir}/finished_at"

  mkdir -p "$shard_dir"
  rm -rf "$sqlite_path" "$mirror_root" "$output_path" "$log_path" "$exit_path" "$finished_path"
  printf 'running\n' >"$state_path"
  date -u +%FT%TZ >"$started_path"

  {
    cd "$REPO_ROOT"
    /usr/bin/time -p go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
      --server "$SERVER" \
      --username "$USERNAME" \
      --password "$MAIL_PASSWORD" \
      --mailbox "$MAILBOX" \
      --since-date "$start" \
      --before-date "$end" \
      --sqlite-path "$sqlite_path" \
      --mirror-root "$mirror_root" \
      --output json >"$output_path" 2>"$log_path"
  }
  local status=$?

  printf '%s\n' "$status" >"$exit_path"
  date -u +%FT%TZ >"$finished_path"
  if [[ "$status" -eq 0 ]]; then
    printf 'done\n' >"$state_path"
  else
    printf 'failed\n' >"$state_path"
  fi

  return "$status"
}

refresh_running_pids() {
  local new_pids=()
  local pid
  for pid in "${RUNNING_PIDS[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      new_pids+=("$pid")
      continue
    fi
    if ! wait "$pid"; then
      FAILED_SHARDS=1
    fi
  done
  RUNNING_PIDS=("${new_pids[@]}")
}

write_shard_manifest
log_manager "starting ${MONTHS}-month backfill with parallel=${PARALLEL} base_dir=${BASE_DIR}"

declare -a RUNNING_PIDS=()
FAILED_SHARDS=0

while IFS=$'\t' read -r shard start end; do
  while true; do
    refresh_running_pids
    if (( ${#RUNNING_PIDS[@]} < PARALLEL )); then
      break
    fi
    sleep 5
  done

  log_manager "launching shard=${shard} since=${start} before=${end}"
  run_shard "$shard" "$start" "$end" &
  RUNNING_PIDS+=("$!")
done <"$manifest"

while (( ${#RUNNING_PIDS[@]} > 0 )); do
  refresh_running_pids
  if (( ${#RUNNING_PIDS[@]} > 0 )); then
    sleep 5
  fi
done

if (( FAILED_SHARDS != 0 )); then
  log_manager "backfill completed with failures"
  exit 1
fi

log_manager "backfill completed successfully"
