---
Title: Run Your First Mirror Sync
Slug: smailnail-mirror-first-sync
Short: Bootstrap a local mirror, inspect the plan, and run the first incremental sync safely.
Topics:
- imap
- mirror
- sqlite
- tutorial
Commands:
- mirror
Flags:
- print-plan
- sqlite-path
- mirror-root
- mailbox
- server
- username
- password
- insecure
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial walks through the first safe use of `smailnail mirror`. The goal is to choose stable local paths, confirm the plan, and then perform an incremental sync that leaves both raw `.eml` files and searchable SQLite rows behind.

## What You Will Build

By the end of this tutorial you will have:

- a local SQLite mirror database,
- a local raw-message tree,
- a successful sync report from the CLI,
- enough context to rerun the command incrementally.

## Prerequisites

- A reachable IMAP server
- Username and password for that server
- A build that includes `sqlite_fts5`
- A target mailbox, usually `INBOX`

If you are testing against the local Docker Dovecot fixture in this repository, expect to add `--insecure` because the fixture uses a self-signed certificate.

## Step 1: Pick Stable Local Paths

The mirror only becomes useful when repeated runs point at the same local database and mirror root. If those paths drift between runs, you will keep creating new snapshots instead of updating the same one.

Choose one database path and one mirror root:

```bash
mkdir -p /tmp/smailnail-demo
```

You do not need to precreate the database file itself, but you should know where it will live.

## Step 2: Print The Plan First

Start with a dry run. This confirms the mailbox scope and local storage targets before the command mutates anything.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --print-plan
```

Review the output row carefully. The important fields are `sqlite_path`, `mirror_root`, `selected_mailbox`, `all_mailboxes`, and `batch_size`.

## Step 3: Run The First Real Sync

Once the plan looks correct, remove `--print-plan` and run the real sync:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw
```

On the first run, expect the command to bootstrap the schema and fetch all mail that falls beyond the initial local checkpoint. On later runs, it should only fetch newly seen UIDs unless you reset state.

The `--log-level info` flag is optional but recommended for first runs. It shows live progress on stderr while the final Glazed output row is still being assembled on stdout.

## Step 4: Inspect The Result

The command emits one summary row. For a first run, the most useful fields are:

- `status`
- `mailboxes_synced`
- `messages_fetched`
- `messages_stored`
- `raw_files_written`

If you want a machine-readable record for a script, ask Glazed for JSON:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  ... \
  --output json
```

## Step 5: Rerun Incrementally

Run the same command again with the same local paths. If nothing changed on the server, the next run should report zero or very few newly fetched messages. That is the expected incremental behavior.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw
```

## Step 6: Expand Scope Deliberately

Only after one mailbox works cleanly should you expand to the full account:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --all-mailboxes \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw
```

`--all-mailboxes` is useful, but it increases run time and storage footprint. Start narrow, then widen scope once the storage layout and credentials are proven.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| The command exits before syncing | The IMAP connection settings are incomplete | Check `--server`, `--username`, `--password`, and `--port` |
| The first run is writing into the wrong place | You skipped the dry-run plan check | Re-run with `--print-plan`, fix the paths, and try again |
| The local Dovecot fixture fails TLS verification | The fixture certificate is self-signed | Add `--insecure` for the fixture run |
| The second run refetches more than expected | You changed paths or reset state | Reuse the same `--sqlite-path` and `--mirror-root`, and avoid `--reset-mailbox-state` unless you mean it |

## See Also

- `smailnail help smailnail-mirror-overview` for the architecture and flag model
- `smailnail help smailnail-mirror-maintenance` for reconcile and rebuild flows
- `smailnail mirror --help` for the full command reference
