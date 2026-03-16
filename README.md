# smailnail

`smailnail` is a Go repository for working with IMAP mailboxes and generated test email.

It currently contains three CLIs:

- `smailnail`: search, fetch, and process mail with a YAML DSL or direct CLI flags
- `mailgen`: generate synthetic email from YAML templates and optionally append it to IMAP
- `imap-tests`: helper commands for creating mailboxes and storing fixture messages

There is now also an initial hosted application binary:

- `smailnaild`: hosted app backend with account CRUD, account tests, mailbox previews, rule CRUD, and rule dry-runs

There is also a dedicated MCP binary for the JavaScript runtime:

- `smailnail-imap-mcp`: exposes `executeIMAPJS` and `getIMAPJSDocumentation`

The repository now also contains an initial reusable JavaScript surface:

- `pkg/services/smailnailjs`: a Go service package for rule parsing/building and JS-friendly result shaping
- `pkg/js/modules/smailnail`: a native `go-go-goja` module exposed as `require("smailnail")`

## Build

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go build ./cmd/smailnail ./cmd/mailgen ./cmd/imap-tests ./cmd/smailnail-imap-mcp ./cmd/smailnaild
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

### `smailnail-imap-mcp`

List the exposed MCP tools:

```bash
go run ./cmd/smailnail-imap-mcp mcp list-tools
```

Start the server over stdio, SSE, or streamable HTTP:

```bash
go run ./cmd/smailnail-imap-mcp mcp start --transport stdio
go run ./cmd/smailnail-imap-mcp mcp start --transport sse --port 3201
go run ./cmd/smailnail-imap-mcp mcp start
```

The default HTTP deployment shape now is:

- transport: `streamable_http`
- port: `3201`

The server intentionally exposes only two tools:

- `executeIMAPJS`: run JavaScript against `require("smailnail")`
- `getIMAPJSDocumentation`: query embedded package/symbol/example/concept docs or render markdown

Production packaging and Coolify deployment notes are in `docs/deployments/smailnail-imap-mcp-coolify.md`. The repository root `Dockerfile` is now the Coolify-facing build entrypoint for this MCP service.

### `smailnaild`

The hosted backend now requires an encryption key for stored IMAP credentials:

```bash
export SMAILNAILD_ENCRYPTION_KEY="$(openssl rand -base64 32)"
```

Start the hosted backend with the default SQLite app database:

```bash
go run ./cmd/smailnaild serve
```

That defaults to:

- bind address `0.0.0.0:8080`
- application DB `smailnaild.sqlite`

Useful endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /api/info`
- `GET /api/accounts`
- `POST /api/accounts`
- `POST /api/accounts/:id/test`
- `GET /api/accounts/:id/mailboxes`
- `GET /api/accounts/:id/messages`
- `GET /api/accounts/:id/messages/:uid`
- `GET /api/rules`
- `POST /api/rules`
- `POST /api/rules/:id/dry-run`

Use Clay SQL flags to point it at another database. For example, Postgres via DSN:

```bash
go run ./cmd/smailnaild serve \
  --listen-host 0.0.0.0 \
  --listen-port 8080 \
  --dsn 'postgres://user:pass@localhost:5432/smailnail?sslmode=disable'
```

Or SQLite with an explicit file path:

```bash
go run ./cmd/smailnaild serve \
  --db-type sqlite \
  --database ./data/smailnaild.sqlite
```

Local hosted-account testing notes and curl examples are in `docs/smailnaild-local-account-flow.md`.

## Environment variables

The Cobra parser is configured with app name `smailnail`, so shared IMAP settings can be supplied with `SMAILNAIL_*` variables such as:

- `SMAILNAIL_SERVER`
- `SMAILNAIL_PORT`
- `SMAILNAIL_USERNAME`
- `SMAILNAIL_PASSWORD`
- `SMAILNAIL_MAILBOX`
- `SMAILNAIL_INSECURE`

The hosted binary uses app name `smailnaild`, so its flags can also be supplied through `SMAILNAILD_*` environment variables.

Important hosted-backend variable:

- `SMAILNAILD_ENCRYPTION_KEY`: base64-encoded 32-byte AES-GCM key for encrypting stored IMAP passwords

## Examples

- DSL examples: `examples/smailnail/*.yaml`
- mailgen configs: `examples/mailgen/*.yaml`
- quick start: `examples/smailnail/QUICK-START.md`

## Docker IMAP fixture

The maintained smoke script looks for the Dovecot fixture in this order:

- `DOCKER_IMAP_FIXTURE_ROOT`
- `../docker-test-dovecot` relative to the `smailnail` repo root

Start it with:

```bash
cd /path/to/docker-test-dovecot
docker compose up -d --build
```

The default test users are `a`, `b`, `c`, and `d`, each with password `pass`.

To run the maintained end-to-end smoke test:

```bash
cd /path/to/smailnail
make smoke-docker-imap
```

If the fixture lives somewhere else locally, override it with `DOCKER_IMAP_FIXTURE_ROOT=/path/to/docker-test-dovecot`.

## Local Dovecot + Keycloak Stack

For hosted-app and OIDC work, the repo now includes a local Docker Compose stack with:

- Dovecot test fixture on the usual local IMAP ports
- Keycloak on `http://127.0.0.1:18080`
- PostgreSQL backing Keycloak persistence

Start it with:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml up -d
```

Useful endpoints and defaults:

- Dovecot IMAPS: `127.0.0.1:993`
- Dovecot test users: `a`, `b`, `c`, `d`
- Dovecot password: `pass`
- Keycloak admin: `http://127.0.0.1:18080/admin`
- Keycloak bootstrap admin username: `admin`
- Keycloak bootstrap admin password: `admin`
- Imported realm: `smailnail-dev`
- Realm issuer: `http://127.0.0.1:18080/realms/smailnail-dev`

The stack also imports two initial OIDC clients in the `smailnail-dev` realm:

- `smailnail-web`
- `smailnail-mcp`

Stop it with:

```bash
docker compose -f docker-compose.local.yml down
```

To run the hosted-backend integration suite against the local Dovecot fixture:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
export SMAILNAILD_ENCRYPTION_KEY="$(openssl rand -base64 32)"
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/...
SMAILNAILD_DOVECOT_TEST=1 go test ./...
```

## JavaScript module smoke

To validate the initial JavaScript service/module slice:

```bash
cd /path/to/smailnail
make smoke-js-module
```

That smoke path runs the service-layer tests and the goja runtime integration tests that prove `require("smailnail")` works.

To validate the dedicated MCP binary and docs registry:

```bash
cd /path/to/smailnail
make smoke-imap-js-mcp
```
