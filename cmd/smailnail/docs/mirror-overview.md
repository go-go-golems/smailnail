---
Title: Local IMAP Mirror Overview
Slug: smailnail-mirror-overview
Short: Understand what `smailnail mirror` stores locally, how sync works, and which flags control the mirror lifecycle.
Topics:
- imap
- mirror
- sqlite
- search
- glazed
Commands:
- mirror
Flags:
- sqlite-path
- mirror-root
- batch-size
- all-mailboxes
- print-plan
- reconcile-full-mailbox
- reset-mailbox-state
- mailbox
- insecure
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

`smailnail mirror` is the durable sync path in this CLI. It connects to IMAP, downloads raw RFC 822 messages, stores those `.eml` files under a local mirror root, and imports parsed metadata and searchable text into SQLite with FTS5 enabled.

## Why Use The Mirror Command

Use `mirror` when you want a local mailbox snapshot that survives after the IMAP session ends. Unlike `fetch-mail`, which streams server results directly to stdout, the mirror command creates a reusable local workspace that is good for repeated searches, later imports, debugging parser behavior, and offline inspection.

The mirror has two durable artifacts:

- A SQLite database, which stores mailbox sync state, parsed headers, searchable text, and reconcile status.
- A raw-message tree, which stores the original downloaded message source as `.eml` files.

That split matters because the SQLite rows are convenient projections, while the raw files remain the canonical source material if parsing improves later.

## What The Command Does

A normal sync run follows this sequence:

1. Open or create the SQLite mirror database.
2. Ensure the mirror schema and FTS tables exist.
3. Connect to the IMAP server with the provided credentials.
4. Select one mailbox or enumerate all mailboxes.
5. Load the stored UID and `UIDVALIDITY` snapshot for each mailbox.
6. Search for new message UIDs and fetch them in batches.
7. Persist raw `.eml` files under the mirror root.
8. Parse the raw messages into searchable SQLite columns.
9. Optionally reconcile the full mailbox and tombstone rows that disappeared remotely.

The command reports these results as a single Glazed output row, which means you can render it as a table, JSON, YAML, or another Glazed output format.

## Important Flags

`--sqlite-path` chooses the SQLite database file. Keep it stable across runs if you want incremental syncs and accumulated history.

`--mirror-root` chooses where raw `.eml` files are written. The command stores files by account, mailbox, `UIDVALIDITY`, and UID so incremental syncs can reuse prior writes safely.

`--batch-size` bounds how many UIDs are fetched per IMAP batch. Increase it if the server is fast and latency dominates. Decrease it if the server is fragile or memory pressure matters.

`--all-mailboxes` changes the scope from one selected mailbox to every listed mailbox the account exposes.

`--print-plan` shows the resolved local storage paths and sync scope without mutating the mirror. Use this before the first real run when you want to verify the target layout.

`--reconcile-full-mailbox` performs a full UID search after incremental sync and marks locally mirrored rows as `remote_deleted` when the server no longer reports them.

`--reset-mailbox-state` clears the stored local sync checkpoint before the run. Use it when the local state is no longer trustworthy or when you intentionally want to rebuild a mailbox snapshot from scratch.

## Build Requirement

The mirror feature requires SQLite FTS5 support at build time. In this repository, the supported path is:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror --help
```

An untagged build that reaches `pkg/mirror` fails on purpose. That is deliberate: the mirror schema now assumes FTS5 is available instead of degrading to a weaker runtime mode.

## Output Fields

The command emits a row with operational fields such as:

- `status`
- `sqlite_path`
- `mirror_root`
- `schema_version`
- `selected_mailbox`
- `all_mailboxes`
- `batch_size`
- `mailboxes_synced`
- `messages_fetched`
- `messages_stored`
- `raw_files_written`
- `messages_tombstoned`
- `messages_restored`

That makes it easy to inspect runs interactively:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror ... --output json
```

## When To Use Reconcile And Reset

Use `--reconcile-full-mailbox` when you care about remote deletions and expunges being reflected in the local mirror. This mode is more expensive because it inspects the full mailbox UID set after the incremental pass, but it gives you a more accurate `remote_deleted` view.

Use `--reset-mailbox-state` when the local checkpoint is wrong, when a test run polluted the database, or when you intentionally want to reimport a mailbox. It does not mean “delete all files first.” It means “forget the saved sync cursor and derive a fresh snapshot on the next run.”

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| Build fails with an FTS5-related error | `sqlite_fts5` or `fts5` build tag was not enabled | Run the command with `-tags sqlite_fts5` |
| Sync keeps skipping mail you expected | Local mailbox state is ahead of what you want to test | Rerun with `--reset-mailbox-state` |
| Deleted remote mail still looks present locally | Reconcile was not requested | Rerun with `--reconcile-full-mailbox` |
| TLS handshake fails against a local fixture | The test server uses a self-signed cert | Use `--insecure` for local fixture runs only |
| Mirror data lands in the wrong directory | `--sqlite-path` or `--mirror-root` resolved differently than expected | Run once with `--print-plan` and check the reported paths |

## See Also

- `smailnail help smailnail-mirror-first-sync` for a step-by-step first mirror run
- `smailnail help smailnail-mirror-maintenance` for reconcile and reset workflows
- `smailnail mirror --help` for the command-level flag reference
- `smailnail help smailnail-mail-app-rules` for the non-mirroring IMAP commands
