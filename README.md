# smailnail

`smailnail` is a Go repository for working with IMAP mailboxes and generated test email.

It currently contains three CLIs:

- `smailnail`: search, fetch, and process mail with a YAML DSL or direct CLI flags
- `mailgen`: generate synthetic email from YAML templates and optionally append it to IMAP
- `imap-tests`: helper commands for creating mailboxes and storing fixture messages

## Build

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go build ./cmd/smailnail ./cmd/mailgen ./cmd/imap-tests
```

## Commands

### `smailnail`

Rule-driven execution:

```bash
go run ./cmd/smailnail mail-rules \
  --rule examples/smailnail/recent-emails.yaml \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --output json
```

Direct fetch via flags:

```bash
go run ./cmd/smailnail fetch-mail \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --subject-contains "invoice" \
  --output json
```

### `mailgen`

```bash
go run ./cmd/mailgen generate \
  --configs examples/mailgen/simple.yaml \
  --write-files \
  --output-dir ./output \
  --output json
```

To append generated mail to IMAP:

```bash
go run ./cmd/mailgen generate \
  --configs examples/mailgen/simple.yaml \
  --store-imap \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --output json
```

### `imap-tests`

Create a mailbox:

```bash
go run ./cmd/imap-tests create-mailbox \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --new-mailbox Scratch \
  --output json
```

Store a text message:

```bash
go run ./cmd/imap-tests store-text-message \
  --server imap.example.com \
  --username user@example.com \
  --password secret \
  --mailbox INBOX \
  --from "Sender <sender@example.com>" \
  --to "Recipient <recipient@example.com>" \
  --subject "Fixture message" \
  --output json
```

## Environment variables

The Cobra parser is configured with app name `smailnail`, so shared IMAP settings can be supplied with `SMAILNAIL_*` variables such as:

- `SMAILNAIL_SERVER`
- `SMAILNAIL_PORT`
- `SMAILNAIL_USERNAME`
- `SMAILNAIL_PASSWORD`
- `SMAILNAIL_MAILBOX`
- `SMAILNAIL_INSECURE`

## Examples

- DSL examples: `examples/smailnail/*.yaml`
- mailgen configs: `examples/mailgen/*.yaml`
- quick start: `examples/smailnail/QUICK-START.md`

## Docker IMAP fixture

The repository is validated against the local Dovecot fixture at:

`/home/manuel/code/others/docker-test-dovecot`

Start it with:

```bash
cd /home/manuel/code/others/docker-test-dovecot
docker compose up -d --build
```

The default test users are `a`, `b`, `c`, and `d`, each with password `pass`.

To run the maintained end-to-end smoke test:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
make smoke-docker-imap
```

If the fixture lives somewhere else locally, override it with `DOCKER_IMAP_FIXTURE_ROOT=/path/to/docker-test-dovecot`.
