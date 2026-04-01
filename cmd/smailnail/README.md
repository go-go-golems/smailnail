# smailnail

`smailnail` is the IMAP DSL and fetch CLI in this repository.

It supports two main flows:

- `mail-rules`: load a YAML rule file and optionally execute actions on matched messages
- `fetch-mail`: build a temporary rule from CLI flags for quick searches
- `mirror`: mirror IMAP mail into a local SQLite database plus raw `.eml` files

## Build

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go build -tags sqlite_fts5 ./cmd/smailnail
```

`smailnail` requires the `sqlite_fts5` build tag so the mirror database can create and query its FTS5 search index.

## Help

```bash
go run -tags sqlite_fts5 ./cmd/smailnail --help
go run -tags sqlite_fts5 ./cmd/smailnail mail-rules --help
go run -tags sqlite_fts5 ./cmd/smailnail fetch-mail --help
go run -tags sqlite_fts5 ./cmd/smailnail mirror --help
```

## Rule-driven usage

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mail-rules \
  --rule examples/smailnail/recent-emails.yaml \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --output json
```

Rules can also include `actions:` blocks for:

- flag changes
- copy
- move
- delete
- export

## Direct fetch usage

```bash
go run -tags sqlite_fts5 ./cmd/smailnail fetch-mail \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --subject-contains "important" \
  --include-content \
  --output json
```

## Shared IMAP flags

Both subcommands accept:

- `--server`
- `--port`
- `--username`
- `--password`
- `--mailbox`
- `--insecure`

These can also be supplied through `SMAILNAIL_*` environment variables.

## Local mirror usage

Bootstrap and sync one mailbox into a local mirror:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --sqlite-path ./smailnail-mirror.sqlite \
  --mirror-root ./smailnail-mirror \
  --output json
```

Mirror all listed mailboxes:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --all-mailboxes \
  --sqlite-path ./smailnail-mirror.sqlite \
  --mirror-root ./smailnail-mirror \
  --output json
```

Print the plan without mutating local state:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --print-plan \
  --output json
```

Reset the stored local checkpoint before a resync:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mirror \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --reset-mailbox-state \
  --sqlite-path ./smailnail-mirror.sqlite \
  --mirror-root ./smailnail-mirror \
  --output json
```

## Examples

- Quick start: `examples/smailnail/QUICK-START.md`
- Rule corpus: `examples/smailnail/*.yaml`
