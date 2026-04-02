---
Title: Mirror Maintenance And Reconciliation
Slug: smailnail-mirror-maintenance
Short: Reconcile remote deletions, reset local mailbox state, and choose the right recovery workflow for an existing mirror.
Topics:
- imap
- mirror
- reconcile
- maintenance
- sqlite
Commands:
- mirror
Flags:
- reconcile-full-mailbox
- reset-mailbox-state
- all-mailboxes
- mailbox-pattern
- exclude-mailbox-pattern
- stop-on-error
- mailbox
- batch-size
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Application
---

This page covers the maintenance workflows you use after a mirror already exists. The two most important operators are reconciliation, which updates `remote_deleted` based on the current server snapshot, and state reset, which tells the mirror to forget its saved mailbox checkpoint and rebuild that mailbox from the server.

## Why Maintenance Exists

Incremental sync is optimized for new mail. That is efficient, but it does not automatically answer every maintenance question. If mail disappears remotely because of deletions or expunges, the local mirror only reflects that once you ask for a full reconcile. If the local checkpoint becomes untrustworthy because of testing or manual DB edits, you need a controlled way to rebuild it.

## Reconcile Remote Deletions

Use `--reconcile-full-mailbox` when you want the local mirror to compare its current snapshot against the server’s full UID set for the mailbox. Any mirrored row missing from that server snapshot is marked `remote_deleted = true`.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --reconcile-full-mailbox
```

Pay attention to `messages_tombstoned` and `messages_restored` in the output row:

- `messages_tombstoned` increases when previously present local rows disappear from the remote mailbox snapshot.
- `messages_restored` increases when a row was previously marked `remote_deleted` but the server still reports it in the current snapshot.

## Rebuild A Mailbox Snapshot

Use `--reset-mailbox-state` when you want to throw away the saved incremental checkpoint for the mailbox before syncing.

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --reset-mailbox-state
```

This is the right recovery tool when:

- you changed test fixtures and want a clean import,
- you suspect the saved local mailbox cursor is wrong,
- you intentionally want to rebuild one mailbox without deleting the whole mirror workspace.

It is not the same as physically deleting the SQLite database or raw files. Resetting state asks the service to compute a new snapshot from the server while keeping the mirror workspace in place.

## Combine Reset And Reconcile Carefully

You can combine both maintenance flags when recovering a mailbox after heavy remote changes:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw \
  --reset-mailbox-state \
  --reconcile-full-mailbox
```

That is more expensive than a routine incremental run, but it gives you the most explicit recovery path short of deleting the mirror artifacts yourself.

## Scope Decisions

Prefer mailbox-scoped maintenance first:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --mailbox Archive \
  ... \
  --reconcile-full-mailbox
```

Only move to `--all-mailboxes` once one mailbox behaves the way you expect. Full-account maintenance can be expensive on large accounts and makes it harder to spot which mailbox caused a surprising result.

When you do widen maintenance to the full account, keep it selective:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --log-level info mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --all-mailboxes \
  --mailbox-pattern 'Archive/*' \
  --exclude-mailbox-pattern 'Archive/Trash*' \
  --stop-on-error=false \
  --reconcile-full-mailbox \
  --sqlite-path /tmp/smailnail-demo/mirror.sqlite \
  --mirror-root /tmp/smailnail-demo/raw
```

This keeps maintenance focused and lets the run finish other mailboxes even if one mailbox returns an IMAP error. Review `mailbox_errors` and `failed_mailboxes` in the output row afterward.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| Remote deletions are not reflected locally | You ran a normal incremental sync without reconcile | Rerun with `--reconcile-full-mailbox` |
| A mailbox now looks duplicated or unexpectedly large | The local checkpoint was reset intentionally or by mistake | Confirm whether `--reset-mailbox-state` was used; if so, review the output counters and local paths |
| Reconcile is slower than expected | It needs a full mailbox UID snapshot | Use it as a maintenance workflow, not every lightweight sync |
| The wrong mailbox was rebuilt | The selected mailbox was broader than intended | Start with `--mailbox <name>` before using `--all-mailboxes` |

## See Also

- `smailnail help smailnail-mirror-overview` for the mirror architecture and storage model
- `smailnail help smailnail-mirror-first-sync` for the initial bootstrap flow
- `smailnail mirror --help` for the raw flag reference
