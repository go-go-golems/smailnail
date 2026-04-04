---
Title: Investigation Report and Implementation Plan
Ticket: SMN-20260403-FTS-SYNC
Status: active
Topics:
    - mirror
    - sqlite
    - fts
    - bug-investigation
DocType: design
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/service.go:resetMailboxState (line 868) is the only mutation path missing FTS cleanup"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/schema.go:Standalone FTS5 table definition (line 142)"
    - "/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/mirror/service_test.go:Needs new test for FTS cleanup on mailbox reset"
ExternalSources: []
Summary: "The GitHub issue's core claim is largely wrong — the main sync path correctly updates messages_fts. However, investigation found a real (lower-severity) bug: resetMailboxState orphans FTS rows."
LastUpdated: 2026-04-03T07:45:00.000000000-04:00
WhatFor: "Document investigation findings and concrete fix plan for FTS synchronization gaps"
WhenToUse: ""
---

# Investigation Report: messages_fts Synchronization on Message Upserts

## Executive Summary

**The GitHub issue's core claim is incorrect.** The normal mirror sync path (`smailnail mirror`) **does** correctly update `messages_fts` alongside `messages` on every upsert. The automated review tool was misled because `upsertMessageRecord` and `upsertMessageFTS` are separate functions — it analyzed the INSERT function in isolation and missed that the caller always pairs them.

**However, the investigation uncovered a real bug:** `resetMailboxState()` deletes messages without cleaning up corresponding `messages_fts` rows, leaving orphaned FTS entries that could produce ghost search results.

## Findings

### 1. Main sync path: ✅ Correct (false positive in the issue)

The `persistMirrorBatch` function (service.go:650-700) correctly calls both functions in sequence within the same transaction:

```go
messageRowID, err := upsertMessageRecord(ctx, tx, record)  // line 678
// ...
if err := upsertMessageFTS(ctx, tx, messageRowID, record); err != nil {  // line 682
    return err
}
```

This is covered by `TestServiceSyncPersistsIncrementalMessages` which asserts both `countFTSRows` and `ftsMatches` after initial and incremental sync.

### 2. Merge path: ✅ Correct (by design)

The merge path (`merge.go:404`) intentionally skips per-row FTS updates for performance, then does a bulk `rebuildMessagesFTS()` (DELETE all + INSERT from messages). This is gated by the `RebuildFTS` flag which defaults to `true` in the CLI.

### 3. `resetMailboxState`: ❌ Bug — orphaned FTS rows

**File:** `pkg/mirror/service.go:868-888`

The function deletes all messages for a given account/mailbox but does not delete corresponding FTS rows:

```go
func resetMailboxState(ctx context.Context, db *sqlx.DB, accountKey, mailboxName string) error {
    tx, err := db.BeginTxx(ctx, nil)
    defer func() { _ = tx.Rollback() }()

    // Deletes messages but NOT messages_fts
    _, err = tx.ExecContext(ctx, `DELETE FROM messages WHERE account_key = ? AND mailbox_name = ?`,
        accountKey, mailboxName)
    _, err = tx.ExecContext(ctx, `DELETE FROM mailbox_sync_state WHERE account_key = ? AND mailbox_name = ?`,
        accountKey, mailboxName)

    return tx.Commit()
}
```

**Triggered by:**
- `--reset-mailbox-state` CLI flag (service.go:343)
- Automatic UIDVALIDITY change detection (service.go:356)

**Impact:** Low-to-moderate. After reset, re-sync populates new FTS rows correctly, but old orphaned FTS rows remain. FTS searches may return rowids pointing to non-existent messages. UIDVALIDITY changes are rare in practice.

### 4. Non-FTS column updates: ✅ Not applicable

The `UPDATE` paths for `remote_deleted`, `thread_id`, `thread_depth`, `sender_email`, and `sender_domain` do not touch FTS-indexed columns, so no FTS update is needed.

### 5. Schema design: standalone FTS table (no automatic sync)

The FTS5 table is defined **without** `content=messages`:

```sql
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    account_key, mailbox_name, subject, from_summary, to_summary,
    cc_summary, body_text, body_html, search_text
)
```

This means SQLite provides **no automatic synchronization**. All sync is the application's responsibility. This is a valid design choice (avoids FTS5 content-sync complications) but requires discipline at every mutation site.

## Why the automated review tool got it wrong

The tool analyzed `upsertMessageRecord` (lines 1026-1030) in isolation and saw:
1. No FTS update inside the function body
2. No SQL trigger defined in the schema

It concluded FTS would be stale. But the **calling code** in `persistMirrorBatch` always pairs `upsertMessageRecord` + `upsertMessageFTS` in the same transaction. The tool didn't trace the call graph.

This is a good lesson in the limits of static analysis on separated functions — the coupling is correct but implicit.

## Implementation Plan

### Fix 1: Add FTS cleanup to `resetMailboxState` (required)

**Severity:** Medium  
**Effort:** Small (< 15 min)

Add a `DELETE FROM messages_fts` statement inside `resetMailboxState`, after deleting from `messages`:

```go
func resetMailboxState(ctx context.Context, db *sqlx.DB, accountKey, mailboxName string) error {
    tx, err := db.BeginTxx(ctx, nil)
    // ...
    if _, err := tx.ExecContext(ctx,
        `DELETE FROM messages WHERE account_key = ? AND mailbox_name = ?`,
        accountKey, mailboxName); err != nil {
        return errors.Wrap(err, "delete mirrored mailbox messages")
    }
    // NEW: clean up orphaned FTS rows
    if _, err := tx.ExecContext(ctx,
        `DELETE FROM messages_fts WHERE account_key = ? AND mailbox_name = ?`,
        accountKey, mailboxName); err != nil {
        return errors.Wrap(err, "delete mirrored mailbox messages_fts rows")
    }
    // ... rest unchanged
}
```

Since `messages_fts` stores `account_key` and `mailbox_name` as indexed columns, this query will work correctly (FTS5 supports column equality matching for DELETE).

### Fix 2: Add test for FTS cleanup on mailbox reset (required)

**Effort:** Small (< 15 min)

Extend `TestServiceSyncResetsOnUIDValidityChange` (or add a new test) to assert:
1. FTS rows exist after initial sync
2. After UIDVALIDITY-triggered reset + re-sync, FTS row count matches messages count (no orphans)
3. FTS does not match subjects from the pre-reset epoch

### Fix 3 (optional): Consider making FTS update atomic with record upsert

**Effort:** Medium  
**Priority:** Low (nice-to-have, not required)

To prevent future callers from forgetting FTS, `upsertMessageRecord` could be refactored to internally call `upsertMessageFTS`:

```go
func upsertMessageRecord(ctx context.Context, tx *sqlx.Tx, record MessageRecord) (int64, error) {
    // ... existing INSERT INTO messages ...
    // ... SELECT id ...
    if err := upsertMessageFTS(ctx, tx, messageRowID, record); err != nil {
        return 0, err
    }
    return messageRowID, nil
}
```

**Trade-off:** The merge path currently calls `upsertMessageRecord` without FTS (for bulk rebuild efficiency). It would need a `skipFTS` parameter or a separate `upsertMessageRecordOnly` function. This adds complexity for marginal safety gain, since the merge path deliberately bulk-rebuilds. **Recommendation: skip this for now**, the two-call pattern is well-established and tested.

### Fix 4 (optional): Add a `doctest`-style invariant check

**Effort:** Small  
**Priority:** Low

Add a helper in tests that asserts `messages` and `messages_fts` row counts match after any sync operation. Could be wired into `openTestStore` as a cleanup hook.

## Recommended Implementation Order

1. **Fix 1** — Add FTS cleanup to `resetMailboxState` (the actual bug)
2. **Fix 2** — Add test coverage for the fix
3. Close the GitHub issue with explanation that the main sync path was already correct, but `resetMailboxState` had a real gap that's now fixed.
4. Optionally pursue Fix 3 or Fix 4 in a follow-up if the codebase gains more mutation paths.

## Summary Table

| Mutation path | FTS updated? | Status |
|---|---|---|
| `persistMirrorBatch` (sync) | ✅ `upsertMessageFTS` called per row | Correct |
| `mergeShards` (merge) | ✅ Bulk `rebuildMessagesFTS` after merge | Correct |
| `resetMailboxState` (reset/UIDVALIDITY) | ❌ No FTS cleanup | **Bug — Fix 1** |
| `updateRemoteDeletedState` | N/A (non-FTS column) | Correct |
| `enrich/threads` | N/A (non-FTS column) | Correct |
| `enrich/senders` | N/A (non-FTS column) | Correct |
