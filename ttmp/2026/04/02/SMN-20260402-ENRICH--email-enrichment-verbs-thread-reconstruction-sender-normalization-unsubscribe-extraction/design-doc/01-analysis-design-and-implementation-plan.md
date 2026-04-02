---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: smailnail/cmd/smailnail/commands/mirror.go
      Note: Mirror command to extend with --enrich-after flag
    - Path: smailnail/cmd/smailnail/main.go
      Note: Root command where enrich group gets registered
    - Path: smailnail/pkg/mailutil/addresses.go
      Note: Existing address parsing utility to supersede with enrichment parser
    - Path: smailnail/pkg/mirror/schema.go
      Note: Migration system to extend with v2 enrichment tables
    - Path: smailnail/pkg/mirror/store.go
      Note: OpenStore/Bootstrap pattern to reuse in enrichment commands
    - Path: smailnail/pkg/mirror/types.go
      Note: MessageRecord struct — the input to all enrichment passes
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Email Enrichment Verbs: Analysis, Design & Implementation Plan

**Ticket:** SMN-20260402-ENRICH  
**Date:** 2026-04-02  
**Scope:** Three post-download enrichment passes over the smailnail SQLite mirror — thread reconstruction, sender normalization, unsubscribe link extraction — implemented as Glazed CLI verbs and integrated optionally into `smailnail mirror`.

---

## Executive Summary

The smailnail SQLite mirror (`pkg/mirror`) stores raw email metadata per-message but makes no attempt to link messages into threads, normalize senders, or surface unsubscribe mechanisms. Three enrichment passes can be added that each add derived tables and columns to the same database file without touching the sync pipeline itself. Each enrichment is idempotent and can be run standalone post-hoc or automatically after `mirror` completes. They are exposed as `smailnail enrich threads`, `smailnail enrich senders`, and `smailnail enrich unsubscribe`, all registered under a new `enrich` Cobra group.

---

## 1. Existing Codebase Analysis

### 1.1 Module & build

- Module: `github.com/go-go-golems/smailnail` (Go 1.26.1)
- SQLite driver: `github.com/mattn/go-sqlite3` (CGO)
- DB access: `github.com/jmoiron/sqlx`
- Schema migrations: hand-rolled in `pkg/mirror/schema.go` via `schemaMigrations()` slice
- Already imported: `github.com/emersion/go-message` (RFC 5322 address/header parsing — covers encoded words, multi-address headers)
- Glazed: `github.com/go-go-golems/glazed v1.0.5`

### 1.2 `messages` table (the input)

Every downloaded message has:

| Column | Type | Notes |
|--------|------|-------|
| `id` | INTEGER PK | |
| `message_id` | TEXT | RFC 5322 `Message-Id` header, stored as-is including `<>` brackets |
| `from_summary` | TEXT | `"Display Name <email@domain>"` — already decoded |
| `headers_json` | TEXT | Full headers as `{"Header-Name": "value", ...}` JSON object |
| `body_text` | TEXT | Plain-text body |
| `sent_date` | TEXT | ISO 8601 with tz |

`headers_json` always contains the keys `In-Reply-To`, `References`, and `List-Unsubscribe` when present in the original message. Keys with no value are absent from the object. 1,686/2,894 messages have `List-Unsubscribe`.

### 1.3 Command pattern

Existing commands (`mirror`, `fetch-mail`, `mail-rules`) all follow the same pattern:

```
pkg/imap/layer.go            ← shared IMAP section definition
cmd/smailnail/commands/      ← one file per verb
cmd/smailnail/main.go        ← cobra root, registers verbs
```

Each verb is a struct embedding `*cmds.CommandDescription`, constructed with `cmds.NewCommandDescription`, sections built with `schema.NewSection` + `fields.New`, and exposes `RunIntoGlazeProcessor(ctx, *values.Values, middlewares.Processor)`.

### 1.4 Schema migration conventions

Migrations live in `pkg/mirror/schema.go` as a slice of `schemaMigration{version int, statements []string}`. `bootstrapSchema` iterates and applies any not-yet-applied versions. The current version is `1`. New enrichment tables will be added in **migration 2** so the enrichment schema is always co-located with the mirror schema — the same `store.Bootstrap()` call ensures they exist.

---

## 2. Design Goals & Constraints

1. **Idempotent** — each pass can be run multiple times safely. Re-running only processes rows not yet enriched, using a watermark in `mirror_metadata`.
2. **Non-destructive** — no existing columns or rows are modified in ways that lose data. Enrichment adds new columns (with `ALTER TABLE … ADD COLUMN IF NOT EXISTS`) and new tables.
3. **Post-hoc & inline** — runs standalone (`smailnail enrich <verb>`) or triggered by `smailnail mirror --enrich-after`.
4. **Glazed-native output** — each verb emits a structured progress/summary row per pass (total processed, new rows written, duration) that can be formatted as JSON/YAML/table by the standard `--output` flag.
5. **No new heavy dependencies** — use stdlib `net/mail` + the already-imported `go-message` library; no regex engine, no external NLP.
6. **Scoped by account/mailbox** — all verbs accept `--account-key` and `--mailbox` to limit scope.

---

## 3. Schema Migration 2

All new schema lives in a single migration so `store.Bootstrap()` is always sufficient.

```sql
-- ─── Thread enrichment ────────────────────────────────────────────────────────

-- thread_id on messages: points to the root message_id of the thread.
-- Empty string = not yet enriched.
ALTER TABLE messages ADD COLUMN thread_id TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN thread_depth INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_messages_thread_id ON messages(thread_id);

-- Thread summary (one row per reconstructed thread)
CREATE TABLE IF NOT EXISTS threads (
    thread_id       TEXT PRIMARY KEY,   -- = root message_id
    subject         TEXT NOT NULL DEFAULT '',
    account_key     TEXT NOT NULL DEFAULT '',
    mailbox_name    TEXT NOT NULL DEFAULT '',
    message_count   INTEGER NOT NULL DEFAULT 0,
    participant_count INTEGER NOT NULL DEFAULT 0,
    first_sent_date TEXT NOT NULL DEFAULT '',
    last_sent_date  TEXT NOT NULL DEFAULT '',
    last_rebuilt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ─── Sender normalization ──────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS senders (
    email                       TEXT PRIMARY KEY,
    display_name                TEXT NOT NULL DEFAULT '',
    domain                      TEXT NOT NULL DEFAULT '',
    -- Apple Hide My Email relay: @privaterelay.appleid.com
    is_private_relay            BOOLEAN NOT NULL DEFAULT FALSE,
    -- Guessed real domain from display name (e.g. "Zillow" → "zillow.com")
    relay_display_domain        TEXT NOT NULL DEFAULT '',
    msg_count                   INTEGER NOT NULL DEFAULT 0,
    first_seen_date             TEXT NOT NULL DEFAULT '',
    last_seen_date              TEXT NOT NULL DEFAULT '',
    -- Populated by enrich-unsubscribe pass
    unsubscribe_mailto          TEXT NOT NULL DEFAULT '',
    unsubscribe_http            TEXT NOT NULL DEFAULT '',
    has_list_unsubscribe        BOOLEAN NOT NULL DEFAULT FALSE,
    last_synced_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_senders_domain ON senders(domain);
CREATE INDEX IF NOT EXISTS idx_senders_is_relay ON senders(is_private_relay);

-- ─── Per-message sender foreign key ───────────────────────────────────────────
-- Denormalized for fast join-free queries
ALTER TABLE messages ADD COLUMN sender_email TEXT NOT NULL DEFAULT '';
ALTER TABLE messages ADD COLUMN sender_domain TEXT NOT NULL DEFAULT '';
```

Migration is applied in `schemaMigrations()` as version 2 with `ALTER TABLE … ADD COLUMN` statements guarded by SQLite's `IF NOT EXISTS` syntax (not directly supported; use `PRAGMA table_info` check or catch "duplicate column" error and proceed).

> **SQLite ADD COLUMN caveat:** SQLite supports `ALTER TABLE … ADD COLUMN` but not `IF NOT EXISTS`. The migration wrapper should catch `"duplicate column name"` errors and treat them as no-ops.

---

## 4. Feature 1 — Thread Reconstruction

### 4.1 Algorithm

RFC 5322 threading uses two headers:
- **`In-Reply-To`**: direct parent message ID (one value)
- **`References`**: space-separated ancestry chain, oldest first, newest last

The `References` header is authoritative. Its first entry is the thread root (for well-behaved senders). Its last entry equals `In-Reply-To`.

**Step 1: Build parent map** (in-memory, O(n))

```
parentOf = map[messageID]parentMessageID
depthOf  = map[messageID]int
```

For each message:
- Parse `References` from `headers_json`
- `parentOf[msg.MessageID] = last entry of References before msg.MessageID`
- If `References` is empty, try `In-Reply-To`
- If both empty: this message is a thread root → `parentOf[msg.MessageID] = ""`

**Step 2: Find roots** (walk up the parent chain)

```go
func findRoot(msgID string, parentOf map[string]string, visited map[string]bool) string {
    for {
        p, ok := parentOf[msgID]
        if !ok || p == "" {
            return msgID  // no parent in our DB = this is the root
        }
        if visited[p] {
            return msgID  // cycle guard
        }
        visited[p] = true
        msgID = p
    }
}
```

**Step 3: Assign thread_id and thread_depth**

Walk every message, call `findRoot`, count hops for depth. Write back to DB in a single transaction with batch `UPDATE messages SET thread_id=?, thread_depth=? WHERE id=?`.

**Step 4: Rebuild `threads` table**

```sql
INSERT OR REPLACE INTO threads
  SELECT
    thread_id,
    MIN(subject),
    account_key,
    mailbox_name,
    COUNT(*) AS message_count,
    COUNT(DISTINCT sender_email) AS participant_count,
    MIN(sent_date) AS first_sent_date,
    MAX(sent_date) AS last_sent_date,
    CURRENT_TIMESTAMP
  FROM messages
  WHERE thread_id != ''
  GROUP BY thread_id, account_key, mailbox_name;
```

### 4.2 GitHub special case

GitHub uses stable, predictable message IDs:
- Thread root: `<owner/repo/issues/N@github.com>` or `<owner/repo/pull/N@github.com>`
- All replies have this as the first entry in `References`

This means the standard algorithm handles GitHub perfectly — no special casing required. The thread root will always resolve to the issue/PR opener.

### 4.3 Incremental mode

`mirror_metadata` stores `threads_enriched_watermark` as a high-water `last_synced_at` timestamp. On each run, only messages with `thread_id = ''` are processed. After enrichment, any thread that gained new messages has its `threads` row rebuilt.

### 4.4 Glazed command: `smailnail enrich threads`

```
Package: cmd/smailnail/commands/enrich/
File:    threads.go
Verb:    "enrich threads"
```

**Flags (in `enrich-threads` section):**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sqlite-path` | string | `smailnail-mirror.sqlite` | Path to mirror DB |
| `--account-key` | string | `""` | Limit to one account (empty = all) |
| `--mailbox` | string | `""` | Limit to one mailbox (empty = all) |
| `--rebuild` | bool | false | Reprocess all messages, not just unenriched ones |
| `--dry-run` | bool | false | Compute but do not write |

**Output row:**

```
messages_processed  int   — messages scanned
threads_created     int   — new entries in threads table  
threads_updated     int   — existing threads that gained messages
elapsed_ms          int
```

### 4.5 Package layout

```
pkg/enrich/
  threads.go       ← ThreadEnricher struct, Enrich(ctx, db, opts) ThreadsReport
  senders.go       ← SenderEnricher struct
  unsubscribe.go   ← UnsubscribeEnricher struct
  schema.go        ← migration 2 statements (imported by mirror/schema.go)
  types.go         ← EnrichReport, EnrichOptions, shared helpers
```

The `mirror/schema.go` `schemaMigrations()` slice grows to include version 2, which references `enrich.SchemaMigrationV2Statements()` — keeping enrichment schema co-located with the enrichment code but bootstrapped by the mirror store.

---

## 5. Feature 2 — Sender Normalization

### 5.1 Parsing `from_summary`

`from_summary` is stored as a decoded RFC 5322 address string: `"Display Name <email@domain>"` or just `"email@domain"`. It is already decoded (no `=?utf-8?Q?...?=` escapes) by the mirror parser, but some exotic encoded entries still appear (e.g. `=?utf-8?Q?The=20Providence=20Athen=C3=A6um?=`).

**Parser strategy:** use `github.com/emersion/go-message/mail.ParseAddressList` (already imported) which handles RFC 2047 encoded words correctly. Fall back to `net/mail.ParseAddress` from stdlib.

```go
func ParseFromSummary(raw string) (email, displayName string, err error) {
    addrs, err := mail.ParseAddressList(raw)
    if err != nil {
        // fallback: treat as bare address
        return strings.TrimSpace(raw), "", nil
    }
    if len(addrs) == 0 {
        return "", "", fmt.Errorf("empty address list")
    }
    return strings.ToLower(addrs[0].Address), addrs[0].Name, nil
}
```

### 5.2 Domain extraction & normalization

```go
func ExtractDomain(email string) string {
    parts := strings.SplitN(email, "@", 2)
    if len(parts) != 2 {
        return ""
    }
    return strings.ToLower(parts[1])
}
```

### 5.3 Apple Private Relay detection

Apple's Hide My Email feature routes mail through `@privaterelay.appleid.com`. The actual sender identity is only available via the display name (e.g. `"Zillow <instant-updates_at_...@privaterelay.appleid.com>"`).

```go
const applePrivateRelayDomain = "privaterelay.appleid.com"

func IsPrivateRelay(domain string) bool {
    return domain == applePrivateRelayDomain
}

// GuessRelayDomain tries to extract a real domain from the display name.
// "Zillow" → "zillow.com", "Domestika" → "domestika.org" (best-effort, lowercase slug)
func GuessRelayDomain(displayName string) string {
    slug := strings.ToLower(strings.TrimSpace(displayName))
    slug = strings.ReplaceAll(slug, " ", "")
    // Return as a best-effort slug; callers can suffix ".com"
    return slug
}
```

For private relay senders, `relay_display_domain` stores the lowercased, slug-ified display name. This is enough to group Zillow vs. Domestika vs. Envato even though their `@` domains are identical noise.

### 5.4 Algorithm

1. SELECT `DISTINCT from_summary, MIN(sent_date), MAX(sent_date), COUNT(*)` from messages (optionally scoped by account/mailbox/not-yet-enriched by checking `sender_email = ''`)
2. For each distinct `from_summary`: parse → (email, displayName, domain, isPrivateRelay, relayDisplayDomain)
3. UPSERT into `senders` table
4. Batch UPDATE `messages SET sender_email=?, sender_domain=? WHERE from_summary=?`

The UPDATE on `messages` is the performance-sensitive step. Batching by `from_summary` (not individual message ID) means at most O(distinct_senders) UPDATE statements ≈ O(1000) rather than O(N messages).

### 5.5 Glazed command: `smailnail enrich senders`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sqlite-path` | string | `smailnail-mirror.sqlite` | |
| `--account-key` | string | `""` | |
| `--mailbox` | string | `""` | |
| `--rebuild` | bool | false | Re-normalize all senders, not just new ones |
| `--dry-run` | bool | false | |
| `--show-private-relay` | bool | false | Emit one output row per private-relay sender |

**Output row (summary mode):**

```
senders_created      int
senders_updated      int
messages_tagged      int
private_relay_count  int
elapsed_ms           int
```

**Output row (per-sender mode, `--output table`):**

Emits one row per sender: `email`, `display_name`, `domain`, `is_private_relay`, `relay_display_domain`, `msg_count`, `first_seen_date`, `last_seen_date`.

---

## 6. Feature 3 — Unsubscribe Link Extraction

### 6.1 RFC 2369 `List-Unsubscribe` format

The header value is a comma-separated list of angle-bracket-delimited URIs:

```
List-Unsubscribe: <mailto:unsub@example.com?subject=unsubscribe>, <https://example.com/unsub?t=abc>
```

Some senders only provide a mailto, some only HTTPS, some both. RFC 8058 adds `List-Unsubscribe-Post: List-Unsubscribe=One-Click` indicating the HTTP URL supports a direct POST (one-click unsubscribe).

### 6.2 Parser

```go
// ParseListUnsubscribe extracts mailto and http URLs from the header value.
// Returns (mailtoURL, httpURL, oneClick bool)
func ParseListUnsubscribe(raw string) (mailto, httpURL string, oneClick bool) {
    // Split on ">,<" or ", <" — find all angle-bracket groups
    re := regexp.MustCompile(`<([^>]+)>`)
    matches := re.FindAllStringSubmatch(raw, -1)
    for _, m := range matches {
        uri := strings.TrimSpace(m[1])
        switch {
        case strings.HasPrefix(uri, "mailto:"):
            mailto = uri
        case strings.HasPrefix(uri, "https://"), strings.HasPrefix(uri, "http://"):
            httpURL = uri
        }
    }
    return
}
```

Note: this is the one place where a `regexp` is justified — the pattern is fixed and well-bounded.

### 6.3 Algorithm

1. SELECT `id, from_summary, headers_json` from messages WHERE `headers_json LIKE '%List-Unsubscribe%'` AND `sender_email != ''` (requires senders pass to have run first, or fall back to parsing from_summary inline)
2. Parse `List-Unsubscribe` and `List-Unsubscribe-Post` from each message's `headers_json`
3. UPSERT into `senders` table: set `unsubscribe_mailto`, `unsubscribe_http`, `has_list_unsubscribe = TRUE`
4. If multiple messages from the same sender have different unsubscribe URLs, prefer the most recent one (max `sent_date`)

No separate table needed — unsubscribe data lives in `senders` directly.

### 6.4 Glazed command: `smailnail enrich unsubscribe`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--sqlite-path` | string | `smailnail-mirror.sqlite` | |
| `--account-key` | string | `""` | |
| `--mailbox` | string | `""` | |
| `--rebuild` | bool | false | |
| `--dry-run` | bool | false | |
| `--emit-links` | bool | false | Output one row per sender with unsubscribe links |

**Summary output row:**

```
senders_with_unsubscribe   int   — unique sender emails that have a link
mailto_links               int   — count with mailto
http_links                 int   — count with http URL
one_click_links            int   — count with RFC 8058 one-click
elapsed_ms                 int
```

**Per-sender output row (`--emit-links`):**

`email`, `display_name`, `domain`, `msg_count`, `unsubscribe_mailto`, `unsubscribe_http`, `one_click`

This output can be piped directly into a shell script that iterates and fires HTTP DELETE/POST requests.

---

## 7. Command Group Structure

New directory layout under `cmd/smailnail/commands/`:

```
cmd/smailnail/commands/
  enrich/
    root.go          ← defines "enrich" cobra.Command group + registers sub-verbs
    threads.go       ← ThreadsCommand (Glazed verb)
    senders.go       ← SendersCommand (Glazed verb)
    unsubscribe.go   ← UnsubscribeCommand (Glazed verb)
```

`enrich/root.go` exports `NewEnrichCommand() (*cobra.Command, error)` which builds the group and wires the three children:

```go
func NewEnrichCommand() (*cobra.Command, error) {
    enrichCmd := &cobra.Command{
        Use:   "enrich",
        Short: "Run enrichment passes over the local mirror database",
    }

    for _, factory := range []func() (cmds.Command, error){
        func() (cmds.Command, error) { return NewThreadsCommand() },
        func() (cmds.Command, error) { return NewSendersCommand() },
        func() (cmds.Command, error) { return NewUnsubscribeCommand() },
    } {
        cmd, err := factory()
        if err != nil {
            return nil, err
        }
        cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd,
            cli.WithParserConfig(cli.CobraParserConfig{AppName: "smailnail"}),
        )
        if err != nil {
            return nil, err
        }
        enrichCmd.AddCommand(cobraCmd)
    }
    return enrichCmd, nil
}
```

`main.go` gains one line:

```go
enrichCmd, err := enrich.NewEnrichCommand()
// ...
rootCmd.AddCommand(enrichCmd)
```

### 7.1 `smailnail enrich all`

A fourth verb that runs threads → senders → unsubscribe in sequence, emitting a combined report. Useful for `--enrich-after` integration.

---

## 8. Integration with `smailnail mirror`

The `mirror` command gets a new flag:

```
--enrich-after   bool   default: false
  Run all enrichment passes (threads, senders, unsubscribe) after a
  successful sync. Passes the same --sqlite-path to each enricher.
```

In `MirrorCommand.RunIntoGlazeProcessor`, after the `service.Sync(...)` call succeeds:

```go
if settings.EnrichAfter {
    enrichReport, err := enrich.RunAll(ctx, settings.SQLitePath, enrich.Options{
        AccountKey:  syncReport.AccountKey,
        MailboxName: settings.Mailbox,
    })
    if err != nil {
        // log warning but don't fail the mirror command
        log.Warn().Err(err).Msg("enrichment failed after sync")
    }
    row.Set("enrich_threads_processed", enrichReport.Threads.MessagesProcessed)
    row.Set("enrich_senders_created", enrichReport.Senders.SendersCreated)
    row.Set("enrich_unsubscribe_found", enrichReport.Unsubscribe.SendersWithUnsubscribe)
}
```

---

## 9. Package Design: `pkg/enrich`

```
pkg/enrich/
  types.go        ← Options, ThreadsReport, SendersReport, UnsubscribeReport, AllReport
  schema.go       ← MigrationV2Statements() []string  (called from pkg/mirror/schema.go)
  threads.go      ← ThreadEnricher, Enrich(ctx, db, opts) (ThreadsReport, error)
  senders.go      ← SenderEnricher, Enrich(ctx, db, opts) (SendersReport, error)
  unsubscribe.go  ← UnsubscribeEnricher, Enrich(ctx, db, opts) (UnsubscribeReport, error)
  all.go          ← RunAll(ctx, sqlitePath, opts) (AllReport, error)
  parse_address.go ← ParseFromSummary, ExtractDomain, IsPrivateRelay, GuessRelayDomain
  parse_headers.go ← ParseListUnsubscribe, GetHeader(headersJSON, key)
```

**`GetHeader`** is a small helper used everywhere:

```go
// GetHeader unmarshals headers_json and returns the named header value.
// Header names are matched case-insensitively.
func GetHeader(headersJSON, name string) string {
    var headers map[string]string
    if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
        return ""
    }
    for k, v := range headers {
        if strings.EqualFold(k, name) {
            return v
        }
    }
    return ""
}
```

---

## 10. Idempotency & Watermarking

Each enricher checks/writes a key in `mirror_metadata`:

| Enricher | Key | Value |
|----------|-----|-------|
| threads | `enrich_threads_at` | RFC 3339 timestamp of last run |
| senders | `enrich_senders_at` | RFC 3339 timestamp |
| unsubscribe | `enrich_unsubscribe_at` | RFC 3339 timestamp |

**Incremental strategy:**
- **threads**: process only messages where `thread_id = ''` (unprocessed). `--rebuild` clears all `thread_id` values first.
- **senders**: process only messages where `sender_email = ''`. `--rebuild` clears `sender_email` first.
- **unsubscribe**: scan all messages with `headers_json LIKE '%List-Unsubscribe%'` but only UPDATE senders rows where `has_list_unsubscribe = FALSE`. `--rebuild` clears `has_list_unsubscribe` first.

---

## 11. Error Handling & Transactions

- Each enricher wraps its writes in a single SQLite transaction. If the run is interrupted, nothing is committed — the next run restarts from scratch for that batch.
- For large databases, threads enricher processes messages in batches of 1,000 (configurable via `--batch-size`) and commits each batch independently. Progress is preserved across interrupts.
- `--dry-run` skips the commit entirely and emits what _would_ have been written.

---

## 12. Testing Strategy

### Unit tests

| File | Test |
|------|------|
| `pkg/enrich/parse_address_test.go` | Round-trip of 20 real `from_summary` values including encoded, private-relay, bare-address cases |
| `pkg/enrich/parse_headers_test.go` | `ParseListUnsubscribe` with mailto-only, http-only, both, RFC 8058 one-click, malformed input |
| `pkg/enrich/threads_test.go` | In-memory graph with known thread structures: linear chain, star (broadcast reply-all), fork, cycle detection |

### Integration tests

Use the existing `pkg/mirror/service_test.go` pattern: spin up a small SQLite DB with a handful of hand-crafted `MessageRecord` rows, run each enricher, assert the output tables.

A test fixture set of ~20 messages can be constructed that covers:
- A GitHub PR thread (predictable References chain)
- A personal reply thread
- Single messages with no References
- Private relay senders
- Messages with both mailto and http unsubscribe links

---

## 13. Implementation Order

1. **`pkg/enrich/schema.go`** — define MigrationV2Statements, wire into `pkg/mirror/schema.go`
2. **`pkg/enrich/types.go`** — Options, Report types  
3. **`pkg/enrich/parse_address.go`** + tests
4. **`pkg/enrich/parse_headers.go`** + tests
5. **`pkg/enrich/senders.go`** + integration test (no dependency on threads)
6. **`pkg/enrich/threads.go`** + integration test (no dependency on senders, but can optionally use `sender_email` for participant count)
7. **`pkg/enrich/unsubscribe.go`** + integration test (requires senders to have run; falls back to parsing `from_summary` inline)
8. **`pkg/enrich/all.go`** — RunAll in order: senders → threads → unsubscribe
9. **`cmd/smailnail/commands/enrich/`** — Glazed verbs + root group
10. **Wire into `main.go`**
11. **Wire `--enrich-after` into `mirror.go`**

Total estimated effort: **3–4 days** for a single developer.

---

## 14. CLI Usage Examples

```bash
# Run all enrichment after mirroring
smailnail mirror --sqlite-path mail.db --enrich-after ...

# Run enrichment post-hoc on an existing DB
smailnail enrich senders --sqlite-path /tmp/smailnail-last-month.sqlite
smailnail enrich threads --sqlite-path /tmp/smailnail-last-month.sqlite
smailnail enrich unsubscribe --sqlite-path /tmp/smailnail-last-month.sqlite

# Show all senders as a table
smailnail enrich senders --sqlite-path mail.db --rebuild --output table

# Show all unsubscribe links, pipe to curl for one-click unsubscribes
smailnail enrich unsubscribe --emit-links --output json \
  | jq -r '.[] | select(.one_click) | .unsubscribe_http' \
  | xargs -I{} curl -s -X POST {}

# Query after enrichment
sqlite3 mail.db "
  SELECT t.thread_id, t.subject, t.message_count, t.participant_count
  FROM threads t
  WHERE t.message_count > 5
  ORDER BY t.last_sent_date DESC;"

sqlite3 mail.db "
  SELECT s.domain, s.msg_count, s.has_list_unsubscribe, s.unsubscribe_http
  FROM senders s
  WHERE s.domain NOT LIKE '%github.com%'
  ORDER BY s.msg_count DESC
  LIMIT 20;"
```

---

## 15. Open Questions

1. **Thread root for external references**: if a message has a References chain pointing to a message not in the local DB (e.g. the original newsletter post that triggered a reply), should the thread root be the missing external ID, or the earliest message we actually have? Current design uses earliest-in-DB as root when external root is absent. This means threads may appear "rooted" at a reply rather than the true start.

2. **Multi-account thread merging**: `thread_id` is not scoped by `account_key` — a thread involving two accounts (e.g. `manuel@bl0rg.net` and `wesen@ruinwesen.com`) will correctly merge. Is this desired? Current design allows it.

3. **Sender table primary key**: using `email TEXT PRIMARY KEY` means if the same person sends from multiple addresses, they are distinct rows. A future "contact" concept could merge them. Out of scope for this ticket.

4. **Private relay reverse-lookup**: `GuessRelayDomain` returns a slug, not a real domain. A small static lookup table (Zillow → zillow.com, Domestika → domestika.org) could improve this. Worth a follow-up ticket.

5. **`--enrich-after` granularity**: currently triggers all three passes. Should each pass be independently toggled (`--enrich-threads`, `--enrich-senders`, `--enrich-unsubscribe`)? Simpler to keep one flag for now.
