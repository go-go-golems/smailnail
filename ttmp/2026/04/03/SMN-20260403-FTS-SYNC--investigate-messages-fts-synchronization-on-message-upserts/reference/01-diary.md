---
Title: Diary
Ticket: SMN-20260403-FTS-SYNC
Status: active
Topics:
    - mirror
    - sqlite
    - fts
    - bug-investigation
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/service.go:Main sync path, upsertMessageRecord, upsertMessageFTS, resetMailboxState"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/schema.go:FTS5 table bootstrap (standalone, no content= sync)"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/merge.go:Merge path with conditional rebuildMessagesFTS"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/service_test.go:Tests for sync + FTS, UIDVALIDITY reset test missing FTS assertion"
ExternalSources: []
Summary: "Investigation of GitHub issue claiming messages_fts is not synchronized during mirror upserts. Finding: the main sync path IS correct, but resetMailboxState has an FTS orphan bug."
LastUpdated: 2026-04-03T07:41:04.701090165-04:00
WhatFor: "Track the investigation of FTS synchronization completeness across all message mutation paths"
WhenToUse: ""
---

# Diary

## Goal

Investigate the GitHub issue (P2) claiming that `messages_fts` is not kept synchronized when messages are upserted during mirror sync, leading to incomplete FTS-backed search results. Determine whether the issue is valid, identify the actual scope of any bugs, and produce an implementation plan.

## Step 1: Identify all messages_fts references in the codebase

The first step was to understand where `messages_fts` is used across the codebase and how FTS rows are created, updated, and deleted.

### Prompt Context

**User prompt (verbatim):** "Address issue pkg/mirror/service.go Comment on lines +1026 to +1030 [...] P2 Badge Keep messages_fts synchronized on message upserts [...]"

**Assistant interpretation:** Investigate whether the `messages_fts` FTS5 table is properly kept in sync when messages are upserted during the normal mirror sync path.

**Inferred user intent:** Validate the GitHub issue, find any actual bugs, and plan fixes.

### What I did

- Searched for all references to `messages_fts` across the codebase:
  ```
  grep -rn "messages_fts" --include="*.go" --include="*.sql"
  ```
- Found references in: `service.go`, `schema.go`, `merge.go`, `service_test.go`, `merge_test.go`, and the CLI command `merge_mirror_shards.go`.

### What I found

Key locations:
1. **`schema.go:142`** — `bootstrapFTS()` creates a **standalone** FTS5 table (no `content=messages` clause), meaning SQLite does NOT automatically sync it.
2. **`service.go:1107`** — `upsertMessageFTS()` function: deletes old FTS row by rowid, then inserts new one.
3. **`service.go:678-683`** — The sync batch loop calls `upsertMessageRecord()` → `upsertMessageFTS()` in sequence. ✅ Correct.
4. **`merge.go:404`** — Merge path calls `upsertMessageRecord()` but NOT `upsertMessageFTS()`. Instead, it relies on a separate `rebuildMessagesFTS()` call gated by the `RebuildFTS` flag (default: true).
5. **`merge.go:643`** — `rebuildMessagesFTS()`: bulk DELETE + INSERT from messages table.

### What worked

- The `grep` approach quickly identified all mutation points.

### What I learned

- The FTS table is **standalone** (not a content-synced FTS5 table). This is a deliberate choice that requires manual maintenance at every mutation point.

## Step 2: Verify the main sync path is correct

I needed to confirm whether the issue's core claim — that the sync path does NOT update FTS — is accurate.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Verify whether the sync path at lines 1026-1030 actually updates FTS.

**Inferred user intent:** Determine if the issue is a false positive for the main sync path.

### What I did

- Read the `persistMirrorBatch` function (service.go:650-700) which is the sync hot path.
- Found lines 678-683:
  ```go
  messageRowID, err := upsertMessageRecord(ctx, tx, record)
  if err != nil { return err }
  if err := upsertMessageFTS(ctx, tx, messageRowID, record); err != nil {
      return err
  }
  ```
- Ran the existing test:
  ```
  go test -tags sqlite_fts5 ./pkg/mirror/ -run TestServiceSyncPersistsIncrementalMessages -v -count=1
  ```
  **Result:** PASS. The test explicitly asserts `countFTSRows` and `ftsMatches` after both initial and incremental sync.

### What worked

- The test confirms the main sync path is correct.

### What I learned

- **The GitHub issue's core claim is wrong for the main sync path.** `upsertMessageFTS` IS called after every `upsertMessageRecord` in the sync hot path. The FTS table is correctly populated during normal `smailnail mirror` runs.

### Why the false positive likely occurred

The `upsertMessageRecord` and `upsertMessageFTS` are separate functions. An automated review tool (the GitHub bot) looked at `upsertMessageRecord` in isolation and saw no FTS update inside it, and no SQL trigger in the schema, so it concluded FTS would be stale. It missed that the **caller** pairs the two calls together.

## Step 3: Audit all message mutation paths for FTS consistency

Even though the main sync path is correct, I audited every SQL statement that mutates the `messages` table to check for FTS gaps.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Comprehensive audit of all message mutation paths.

**Inferred user intent:** Find any actual FTS bugs even if the flagged path is fine.

### What I did

Searched for all INSERT/UPDATE/DELETE on `messages`:
```
grep -rn "INSERT INTO messages\|UPDATE messages\|DELETE FROM messages" pkg/ --include="*.go" | grep -v _test.go | grep -v messages_fts
```

Found 5 mutation sites:

| Location | Operation | FTS-indexed columns affected? | FTS updated? | Bug? |
|---|---|---|---|---|
| `service.go:1030` (`upsertMessageRecord`) | INSERT/UPSERT | Yes (subject, from_summary, etc.) | Yes — caller at line 682 calls `upsertMessageFTS` | ✅ No |
| `service.go:877` (`resetMailboxState`) | DELETE FROM messages | Yes — removes rows entirely | **No** — does not delete from `messages_fts` | ❌ **Bug** |
| `service.go:1009` (`updateRemoteDeletedState`) | UPDATE messages SET remote_deleted | No (remote_deleted is not FTS-indexed) | N/A | ✅ No |
| `enrich/threads.go:140` | UPDATE messages SET thread_id, thread_depth | No (not FTS-indexed columns) | N/A | ✅ No |
| `enrich/senders.go:185,266` | UPDATE messages SET sender_email, sender_domain | No (not FTS-indexed columns) | N/A | ✅ No |

### The actual bug: `resetMailboxState` orphans FTS rows

`resetMailboxState` (service.go:868-888) deletes all messages for a mailbox but does NOT delete corresponding `messages_fts` rows:

```go
func resetMailboxState(ctx context.Context, db *sqlx.DB, accountKey, mailboxName string) error {
    tx, err := db.BeginTxx(ctx, nil)
    // ...
    if _, err := tx.ExecContext(ctx, `DELETE FROM messages WHERE account_key = ? AND mailbox_name = ?`, accountKey, mailboxName); err != nil {
        return errors.Wrap(err, "delete mirrored mailbox messages")
    }
    // NO corresponding: DELETE FROM messages_fts WHERE account_key = ? AND mailbox_name = ?
    // ...
}
```

This function is called in two scenarios:
1. **User requests `--reset-mailbox-state`** (service.go:343)
2. **UIDVALIDITY change detected** (service.go:356) — the server reassigned UIDs, so old messages must be discarded.

After reset, the subsequent sync will re-populate `messages` and `messages_fts` with the new data, but the **old** orphaned FTS rows remain. This means:
- FTS searches could return rowids that no longer exist in `messages`
- FTS row count inflates over time with each UIDVALIDITY change

### What was tricky to build

N/A — investigation only.

### What warrants a second pair of eyes

- The `resetMailboxState` FTS orphan bug is confirmed but has **low practical severity**: UIDVALIDITY changes are rare, and manual resets are intentional (the user would likely re-sync immediately). Ghost FTS results would reference non-existent message rowids, which would cause empty/error results rather than wrong data.
- The merge path intentionally skips per-row FTS and does a bulk rebuild — this is correct and more efficient.

### What should be done in the future

See implementation plan below.

### Code review instructions

- **Start at:** `pkg/mirror/service.go:868` — `resetMailboxState` function
- **Verify FTS gap:** Note absence of any `DELETE FROM messages_fts` in that function
- **Compare with:** `pkg/mirror/service.go:678-683` — the sync path which correctly pairs upsertMessageRecord + upsertMessageFTS
- **Run test:** `go test -tags sqlite_fts5 ./pkg/mirror/ -run TestServiceSync -v -count=1`

### Technical details

**FTS5 schema** (schema.go:142):
```sql
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    account_key, mailbox_name, subject, from_summary, to_summary,
    cc_summary, body_text, body_html, search_text
)
```
Note: No `content=messages` — this is a standalone FTS table requiring manual sync.

**FTS columns indexed:** account_key, mailbox_name, subject, from_summary, to_summary, cc_summary, body_text, body_html, search_text.

**Messages columns NOT in FTS:** id, uidvalidity, uid, message_id, internal_date, sent_date, size_bytes, flags_json, headers_json, parts_json, raw_path, raw_sha256, has_attachments, remote_deleted, first_seen_at, last_synced_at, thread_id, thread_depth, sender_email, sender_domain.
