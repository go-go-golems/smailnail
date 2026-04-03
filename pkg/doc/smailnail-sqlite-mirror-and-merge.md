---
Title: "Mirror Mail Into SQLite, Query It, And Merge Shards"
Slug: "smailnail-sqlite-mirror-and-merge"
Short: "Learn the full local-mirror workflow: download IMAP mail into SQLite, inspect and query the database directly, and merge month-sharded mirrors back into one durable local mirror."
Topics:
- imap
- mirror
- sqlite
- search
- cli
Commands:
- mirror
- merge-mirror-shards
Flags:
- sqlite-path
- mirror-root
- since-days
- since-date
- before-date
- reconcile-full-mailbox
- enrich-after
- input-root
- output-sqlite
- output-mirror-root
- shard-glob
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

This page explains the complete local SQLite workflow in `smailnail`. It is for users who want to do more than fetch one mailbox snapshot once. The goal is to understand how to mirror mail into a durable local database, how to inspect and query that SQLite database directly with standard tools, and how to merge month-sharded backfills into one reusable local mirror.

If you are new to this CLI, start here instead of reading the command help pages in isolation. The individual command flags make much more sense once you see the whole lifecycle from server download to local query to shard merge.

## Why This Workflow Exists

The `mirror` command solves a different problem than `fetch-mail`.

- `fetch-mail` is for direct server reads where the output goes straight to stdout.
- `mirror` is for building a durable local workspace that survives after the IMAP session ends.
- `merge-mirror-shards` is for consolidating many bounded mirror runs, usually month slices, into one SQLite database and one raw-message tree.

That distinction matters because a local mirror gives you three things at once:

1. A reusable SQLite database with parsed headers, text, HTML, attachment metadata, and checkpoint state.
2. A raw-message tree containing the original `.eml` source files.
3. A workflow that can be resumed, reconciled, enriched, or merged later instead of starting from zero every time.

If you skip the mirror and only fetch mail directly, you lose that durable local state. If you create month-sharded mirrors but never merge them, you end up with many isolated local databases instead of one long-lived mirror you can keep querying and updating.

## What The Local Mirror Actually Stores

A local mirror has two durable artifacts:

1. A SQLite database, usually something like `smailnail-mirror.sqlite`.
2. A mirror root directory, usually something like `smailnail-mirror/raw/...`.

The SQLite database stores the convenient projection. The raw tree stores the canonical downloaded message source.

### SQLite tables that matter most

These are the tables new users should understand first:

| Table | What it stores | Why it matters |
| --- | --- | --- |
| `messages` | One row per mirrored message | This is the main local mail table |
| `mailbox_sync_state` | Checkpoint state per mailbox | This controls incremental resume behavior |
| `messages_fts` | FTS5 search index rows | This makes local full-text search practical |

There are also enrichment tables once enrichment migrations are present, but those are derived state. The first things to understand are still `messages`, `mailbox_sync_state`, and `messages_fts`.

### Raw files that matter most

The raw-message tree stores `.eml` files under a deterministic relative path:

```text
raw/<account-key>/<mailbox-slug>/<uidvalidity>/<uid>.eml
```

That path is important for two reasons:

1. It keeps the raw files stable across reruns.
2. It lets shard merges copy raw files into a new destination mirror root without rewriting every message row.

## Part 1: Mirror Mail Into SQLite

This section covers the normal download path. It explains how the `mirror` command works in practice and how to choose the right flags for safe first runs.

### The simplest safe mirror run

Start with one mailbox and a bounded date range.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
  --server imap.example.com \
  --username user@example.com \
  --password "$MAIL_PASSWORD" \
  --mailbox INBOX \
  --since-days 30 \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --output json
```

This command does several things:

1. Creates or opens `/tmp/smailnail-demo/mirror.sqlite`.
2. Ensures the mirror schema and FTS tables exist.
3. Connects to the IMAP server.
4. Searches for matching messages in `INBOX`.
5. Downloads raw message source into the mirror root.
6. Stores parsed message projections in SQLite.
7. Emits a structured report.

### Why stable local paths matter

Use the same `--sqlite-path` and `--mirror-root` for repeated runs if you want an incremental mirror.

If you change these paths each time:

- you create new unrelated local databases,
- you lose incremental resume behavior,
- you make later merge or reconciliation work harder.

Think of these paths as the identity of your local mirror.

### Recommended first-run flags

For new users, these flags reduce risk the most:

| Flag | Use it when | Why it helps |
| --- | --- | --- |
| `--since-days 30` | You are testing a real account | Prevents accidental full-history imports |
| `--max-messages 500` | You want a small first smoke test | Caps work even if the date range is larger than expected |
| `--print-plan` | You want to inspect paths and options without mutation | Shows intent before download |
| `--log-level info` | You suspect the sync is slow | Emits batch and mailbox progress logs |
| `--insecure` | You are using the local Docker Dovecot fixture or another self-signed server | Allows TLS to succeed in test environments |

### Plan first, then run

If you want a no-mutation check first:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password "$MAIL_PASSWORD" \
  --mailbox INBOX \
  --since-days 30 \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --print-plan \
  --output json
```

That is especially useful before large historical imports.

### Enrich immediately after download

If you want sender, thread, and unsubscribe enrichment to run right after a successful sync, add `--enrich-after`.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password "$MAIL_PASSWORD" \
  --mailbox INBOX \
  --since-days 30 \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --enrich-after \
  --output json
```

Use this when the destination SQLite DB will be queried immediately and you want derived sender or thread information available right away.

## Part 2: Query The SQLite Mirror Directly

This section covers the most practical SQLite inspection and query patterns. It matters because today the mirror database is designed to be inspectable with ordinary SQLite tooling, not only with custom application code.

### Open the database

Use the standard `sqlite3` CLI:

```bash
sqlite3 /tmp/smailnail-demo/mirror.sqlite
```

If you only want one-off queries, pass them directly:

```bash
sqlite3 /tmp/smailnail-demo/mirror.sqlite 'select count(*) from messages;'
```

### The first queries new users should run

These tell you whether the mirror is healthy.

#### How many messages do I have?

```bash
sqlite3 /tmp/smailnail-demo/mirror.sqlite \
  'select count(*) from messages;'
```

#### What mailboxes are present?

```bash
sqlite3 /tmp/smailnail-demo/mirror.sqlite \
  'select mailbox_name, count(*) from messages group by mailbox_name order by mailbox_name;'
```

#### What is my current incremental checkpoint?

```bash
sqlite3 /tmp/smailnail-demo/mirror.sqlite \
  'select account_key, mailbox_name, uidvalidity, highest_uid, last_uidnext, last_sync_at from mailbox_sync_state order by mailbox_name;'
```

This query matters because `mailbox_sync_state` is what the mirror uses to know how far it got.

#### Which rows were marked as deleted remotely?

```bash
sqlite3 /tmp/smailnail-demo/mirror.sqlite \
  'select mailbox_name, uid, subject, remote_deleted from messages where remote_deleted = 1 order by mailbox_name, uid limit 50;'
```

That query becomes useful once you start using `--reconcile-full-mailbox`.

### Inspect the shape of one message row

The `messages` table is intentionally rich. A useful first peek is:

```bash
sqlite3 -header -column /tmp/smailnail-demo/mirror.sqlite '
  select
    mailbox_name,
    uid,
    subject,
    from_summary,
    sent_date,
    has_attachments,
    raw_path
  from messages
  order by sent_date desc
  limit 20;
'
```

This tells you what the mirror knows locally without reopening IMAP.

### Query parsed headers and text

The `messages` table keeps parsed searchable content such as:

- `subject`
- `from_summary`
- `to_summary`
- `cc_summary`
- `body_text`
- `body_html`
- `search_text`

That means many simple local queries do not need FTS at all.

#### Example: recent messages from one sender

```bash
sqlite3 -header -column /tmp/smailnail-demo/mirror.sqlite "
  select sent_date, subject, from_summary
  from messages
  where from_summary like '%newsletter@example.com%'
  order by sent_date desc
  limit 20;
"
```

#### Example: rows with attachments

```bash
sqlite3 -header -column /tmp/smailnail-demo/mirror.sqlite '
  select sent_date, subject, from_summary, size_bytes
  from messages
  where has_attachments = 1
  order by sent_date desc
  limit 20;
'
```

### Query the FTS index

The `messages_fts` table exists for local full-text search. In practical terms, use it when you want content-style search over subject and bodies instead of exact-match filtering.

#### Example: search for a phrase in the mirror

```bash
sqlite3 -header -column /tmp/smailnail-demo/mirror.sqlite "
  select
    m.mailbox_name,
    m.uid,
    m.sent_date,
    m.subject,
    m.from_summary
  from messages_fts f
  join messages m on m.id = f.rowid
  where messages_fts match 'invoice OR receipt'
  order by m.sent_date desc
  limit 20;
"
```

#### Example: search only one mailbox

```bash
sqlite3 -header -column /tmp/smailnail-demo/mirror.sqlite "
  select
    m.uid,
    m.sent_date,
    m.subject
  from messages_fts f
  join messages m on m.id = f.rowid
  where messages_fts match 'project AND deadline'
    and m.mailbox_name = 'INBOX'
  order by m.sent_date desc
  limit 20;
"
```

### Why query SQLite directly at all?

This matters for new users because it changes how you think about the tool.

Once mail is mirrored locally:

- many debugging questions are SQL questions, not IMAP questions,
- many reporting tasks become one-liner `sqlite3` queries,
- the raw `.eml` tree remains available for deep inspection if a row looks wrong.

That is exactly why the mirror workflow exists.

## Part 3: Reconcile And Maintain The Local Mirror

This section covers the lifecycle after the first import. A local mirror is not just “download once and forget.” It has maintenance operations.

### Reflect remote deletions

Use `--reconcile-full-mailbox` if you want the local mirror to check whether mirrored rows still exist on the server.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password "$MAIL_PASSWORD" \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --reconcile-full-mailbox \
  --output json
```

This does not physically delete rows. It marks them with `remote_deleted = true`, which is usually the safer operator model.

### Rebuild a mailbox checkpoint

Use `--reset-mailbox-state` when the saved checkpoint is no longer trustworthy or you intentionally want that mailbox to be rebuilt.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password "$MAIL_PASSWORD" \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --reset-mailbox-state \
  --output json
```

Use this carefully. It is for rebuilding local state, not for normal incremental operation.

## Part 4: Run Month-Sharded Backfills

This section explains why sharding exists and when to use it.

For large historical imports, a single long-running mirror process can be slow or operationally awkward. The current project already supports bounded mirror runs using:

- `--since-date`
- `--before-date`

That makes it possible to backfill one month at a time in parallel.

### One bounded month run

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password "$MAIL_PASSWORD" \
  --mailbox INBOX \
  --since-date 2026-03-01 \
  --before-date 2026-04-01 \
  --sqlite-path /tmp/backfill/2026-03/mirror.sqlite \
  --mirror-root /tmp/backfill/2026-03/raw \
  --output json
```

If you repeat that for several month windows, you end up with one shard directory per month.

### Why shard by month?

Month sharding works well because it is:

- easy to reason about,
- easy to parallelize,
- easy to inspect later,
- naturally aligned with date-bound sync flags.

The cost is that you do not yet have one unified mirror. That is what the merge command solves.

## Part 5: Merge Shard Databases Back Into One Mirror

This section covers the other half of the SQLite workflow: after many shard-local downloads, how do you get back to one durable local mirror?

Use `merge-mirror-shards`.

### Dry-run first

Always inspect the shard set before mutating the destination.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail merge-mirror-shards \
  --input-root /tmp/backfill \
  --output-sqlite /tmp/merged/mirror.sqlite \
  --output-mirror-root /tmp/merged/root \
  --dry-run \
  --output json
```

This tells you:

- which shard directories were discovered,
- how many messages are in each shard,
- which mailboxes and UIDVALIDITY values are present,
- whether any shard raw roots are missing,
- whether there are immediate mergeability problems.

### Real merge

```bash
go run -tags sqlite_fts5 ./cmd/smailnail merge-mirror-shards \
  --input-root /tmp/backfill \
  --output-sqlite /tmp/merged/mirror.sqlite \
  --output-mirror-root /tmp/merged/root \
  --output json
```

The merge command:

1. discovers shard directories under `--input-root`,
2. inspects each shard DB,
3. bootstraps a fresh destination SQLite DB and mirror root,
4. copies canonical message rows into the destination,
5. copies or reuses raw `.eml` files,
6. rebuilds `mailbox_sync_state`,
7. rebuilds `messages_fts`.

### Enrich after merge

If you want the merged DB enriched immediately:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail merge-mirror-shards \
  --input-root /tmp/backfill \
  --output-sqlite /tmp/merged/mirror.sqlite \
  --output-mirror-root /tmp/merged/root \
  --enrich-after \
  --output json
```

This mirrors the behavior of `smailnail mirror --enrich-after`: the canonical merge happens first, then enrichment runs as a post-step.

### Missing raw files: warning vs strict mode

By default, a missing source raw file produces a warning and increments the missing-raw counter in the report. That behavior is useful when a shard directory is slightly incomplete but the bulk of the DB is still worth merging.

If you want strict behavior instead:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail merge-mirror-shards \
  --input-root /tmp/backfill \
  --output-sqlite /tmp/merged/mirror.sqlite \
  --output-mirror-root /tmp/merged/root \
  --fail-on-missing-raw \
  --output json
```

Use strict mode when you care more about archival integrity than forward progress.

### Inspect the merged DB

After a merge, validate the destination:

```bash
sqlite3 /tmp/merged/mirror.sqlite '
  select count(*) from messages;
  select count(*) from messages_fts;
  select mailbox_name, highest_uid, last_uidnext from mailbox_sync_state;
'
```

These are the first three checks to run because they confirm:

1. the canonical row count,
2. the rebuilt FTS row count,
3. the rebuilt incremental checkpoints.

## Recommended Workflows

This section turns the pieces above into end-to-end recipes.

### Workflow A: Small local mirror for one mailbox

Use this when you are just getting started.

1. Run `mirror` with `--since-days 30`.
2. Query `messages` and `messages_fts` with `sqlite3`.
3. Rerun the same `mirror` command later with the same paths for incremental sync.

### Workflow B: Large history import with month shards

Use this when you want to parallelize a multi-month or multi-year import.

1. Run one month-bounded `mirror` per shard with unique `--sqlite-path` and `--mirror-root`.
2. Keep all shard directories under one parent root.
3. Run `merge-mirror-shards --dry-run`.
4. Run `merge-mirror-shards` for the real merge.
5. Optionally add `--enrich-after`.
6. Query the merged SQLite DB directly.

### Workflow C: Ongoing durable mirror after a historical backfill

Use this when you want to turn the merged DB into the long-lived local mirror.

1. Perform the month-sharded backfill.
2. Merge into one destination mirror.
3. Keep using that merged `--sqlite-path` and `--mirror-root` for future incremental `mirror` runs.

This workflow is the reason the merge feature exists.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `fts5 is required but unavailable` | The binary was not built with the SQLite FTS5 tag | Rebuild or run with `-tags sqlite_fts5` |
| The mirror seems to hang | The first sync is large or the server is slow | Add `--log-level info`, narrow the scope with `--since-days` or `--max-messages`, and verify server reachability |
| `before-date must be after since-date` | The date range is reversed or empty | Use a strictly increasing `[since-date, before-date)` window |
| Incremental sync does not pick up older history after a short test run | The saved checkpoint already advanced past those old messages | Use a fresh `--sqlite-path` or `--reset-mailbox-state` before broadening history |
| `output-sqlite path already exists` during merge | The merge command expects a fresh destination unless overwrite is allowed | Pick a fresh destination path or explicitly use the overwrite flag if that is the intended workflow |
| Merge warns about missing raw files | A shard row points at a raw file that is not present in the shard raw tree | Inspect the warning count; rerun in strict mode with `--fail-on-missing-raw` if integrity is more important than partial progress |
| Merge fails on UIDVALIDITY conflicts | Two shards represent different mailbox generations for the same mailbox | Treat that as a real data-model conflict and inspect the shard set before merging |
| `remote_deleted` rows still appear in local queries | Reconciliation marks rows, it does not physically delete them | Add `where remote_deleted = 0` to user-facing queries when you only want currently present remote mail |

## See Also

- `smailnail help smailnail-mirror-overview` for the original mirror architecture overview
- `smailnail help smailnail-mirror-first-sync` for the step-by-step first mirror tutorial
- `smailnail help smailnail-mirror-maintenance` for reconcile and reset workflows
- `smailnail help smailnail-mail-app-rules` for the non-mirroring IMAP commands
