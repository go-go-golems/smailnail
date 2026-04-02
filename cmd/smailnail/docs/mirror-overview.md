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
- max-messages
- since-days
- all-mailboxes
- mailbox-pattern
- exclude-mailbox-pattern
- stop-on-error
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

`--max-messages` caps the total number of imported messages across the whole run. Use it for the first real sync against a large account when you want proof that storage, parsing, and paths all look correct before you import the full mailbox history.

`--since-days` limits IMAP search to messages newer than the computed cutoff. This is the safest way to start with “recent mail only” instead of mirroring years of history on the first attempt.

`--mailbox-pattern` narrows `--all-mailboxes` to glob-matching mailbox names such as `Archive/*` or `Projects/*`.

`--exclude-mailbox-pattern` removes matching mailbox names from the resolved mailbox list. This is useful for `Trash`, `Spam`, or archival trees you do not want in the first pass.

`--stop-on-error` controls failure semantics during multi-mailbox runs. Leave it at the default `true` when you want the run to abort on the first mailbox failure. Set `--stop-on-error=false` when you want the mirror to keep syncing the remaining mailboxes and report partial completion.

`--print-plan` shows the resolved local storage paths and sync scope without mutating the mirror. Use this before the first real run when you want to verify the target layout.

`--reconcile-full-mailbox` performs a full UID search after incremental sync and marks locally mirrored rows as `remote_deleted` when the server no longer reports them.

`--reset-mailbox-state` clears the stored local sync checkpoint before the run. Use it when the local state is no longer trustworthy or when you intentionally want to rebuild a mailbox snapshot from scratch.

## Progress Logging

Mirror emits its summary row only after the sync finishes, so a large mailbox can look idle if you only watch stdout. Use the root logging flags when you want live progress on stderr:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX
```

`--log-level info` shows high-level progress such as mailbox selection, UID discovery, batch fetches, reconcile passes, and final totals. `--log-level debug` adds lower-level details that are more useful when debugging a specific sync problem.

## Scope And Safety Recipes

For a cautious first sync against a real account, combine a recent-mail cutoff with a hard message cap:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --since-days 30 \
  --max-messages 200
```

For a full-account run that skips noisy folders and keeps going when one mailbox fails:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --all-mailboxes \
  --mailbox-pattern 'Archive/*' \
  --exclude-mailbox-pattern 'Archive/Spam*' \
  --stop-on-error=false \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw
```

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
- `mailbox_pattern`
- `exclude_mailbox_pattern`
- `batch_size`
- `max_messages`
- `since_days`
- `stop_on_error`
- `mailboxes_synced`
- `mailbox_errors`
- `failed_mailboxes`
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
| Full-account runs are too broad | You mirrored every listed mailbox by default | Add `--mailbox-pattern` and `--exclude-mailbox-pattern` before rerunning |
| The run ends with `status=partial` | `--stop-on-error=false` let the run continue after one or more mailbox failures | Inspect `mailbox_errors` and `failed_mailboxes`, then rerun the failed mailboxes directly |
| TLS handshake fails against a local fixture | The test server uses a self-signed cert | Use `--insecure` for local fixture runs only |
| Mirror data lands in the wrong directory | `--sqlite-path` or `--mirror-root` resolved differently than expected | Run once with `--print-plan` and check the reported paths |

## See Also

- `smailnail help smailnail-mirror-first-sync` for a step-by-step first mirror run
- `smailnail help smailnail-mirror-maintenance` for reconcile and reset workflows
- `smailnail mirror --help` for the command-level flag reference
- `smailnail help smailnail-mail-app-rules` for the non-mirroring IMAP commands
