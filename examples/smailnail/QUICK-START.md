# Quick Start

This quick start uses the local Dovecot test fixture and the current `smailnail` CLI layout.

## Prerequisites

- Go installed locally
- Docker available
- The Dovecot fixture repo at `/home/manuel/code/others/docker-test-dovecot`

## Step 1: Start the IMAP fixture

```bash
cd /home/manuel/code/others/docker-test-dovecot
docker compose up -d --build
```

Fixture credentials:

- username: `a`
- password: `pass`
- IMAPS server: `localhost:993`

## Step 2: Build or run the CLI

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go build -tags sqlite_fts5 ./cmd/smailnail
```

You can also use `go run -tags sqlite_fts5 ./cmd/smailnail ...` directly in the examples below.

## Step 3: Fetch recent mail

```bash
go run -tags sqlite_fts5 ./cmd/smailnail fetch-mail \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --insecure \
  --output json
```

## Step 4: Run a YAML rule

```bash
go run -tags sqlite_fts5 ./cmd/smailnail mail-rules \
  --rule examples/smailnail/recent-emails.yaml \
  --server localhost \
  --username a \
  --password pass \
  --mailbox INBOX \
  --insecure \
  --output json
```

## Step 5: Inspect other examples

Useful rule files:

- `examples/smailnail/recent-emails.yaml`
- `examples/smailnail/important-emails.yaml`
- `examples/smailnail/date-range-search.yaml`
- `examples/smailnail/full-message-content.yaml`
- `examples/smailnail/mime-parts-with-content.yaml`

## Notes

- The fixture uses a self-signed certificate, so `--insecure` is expected in local testing.
- Shared IMAP settings can also be provided via environment variables such as `SMAILNAIL_SERVER`, `SMAILNAIL_USERNAME`, and `SMAILNAIL_PASSWORD`.
