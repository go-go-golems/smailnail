# smailnaild local account flow

This document describes the current hosted-backend workflow for `smailnaild` when testing locally against the bundled Dovecot fixture.

## Prerequisites

Start the local Dovecot fixture:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml up -d dovecot
```

Set an application encryption key before starting `smailnaild`:

```bash
export SMAILNAILD_ENCRYPTION_KEY_BASE64="$(openssl rand -base64 32)"
```

This is the Glazed-backed environment form of `--encryption-key-base64`.

The key encrypts stored IMAP passwords inside the application database. If the key changes, previously stored account secrets can no longer be decrypted.

## Start the hosted backend

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go run ./cmd/smailnaild serve \
  --encryption-key-base64 "$SMAILNAILD_ENCRYPTION_KEY_BASE64"
```

Default behavior:

- bind address: `0.0.0.0:8080`
- application DB: `smailnaild.sqlite`
- default user identity for local development: `local-user`
- override per request with `X-Smailnail-User-ID`

## Health and service metadata

```bash
curl -s http://127.0.0.1:8080/healthz | jq
curl -s http://127.0.0.1:8080/readyz | jq
curl -s http://127.0.0.1:8080/api/info | jq
```

## Create a local Dovecot account

```bash
curl -s http://127.0.0.1:8080/api/accounts \
  -H 'Content-Type: application/json' \
  -d '{
    "label": "Local Dovecot",
    "providerHint": "local",
    "server": "localhost",
    "port": 993,
    "username": "a",
    "password": "pass",
    "mailboxDefault": "INBOX",
    "insecure": true,
    "authKind": "password"
  }' | jq
```

The response contains the generated account ID and never echoes the plaintext password.

## Run an account test

```bash
curl -s http://127.0.0.1:8080/api/accounts/<account-id>/test \
  -H 'Content-Type: application/json' \
  -d '{}' | jq
```

The read-only test performs:

- TLS connection
- login
- mailbox selection
- mailbox listing
- sample fetch

## List mailboxes and preview messages

```bash
curl -s http://127.0.0.1:8080/api/accounts/<account-id>/mailboxes | jq
curl -s 'http://127.0.0.1:8080/api/accounts/<account-id>/messages?mailbox=INBOX&limit=20&offset=0' | jq
curl -s 'http://127.0.0.1:8080/api/accounts/<account-id>/messages/<uid>?mailbox=INBOX' | jq
```

Useful query parameters for `GET /api/accounts/<account-id>/messages`:

- `mailbox`
- `limit`
- `offset`
- `query`
- `unread_only`

## Create a rule and dry-run it

```bash
curl -s http://127.0.0.1:8080/api/rules \
  -H 'Content-Type: application/json' \
  -d '{
    "imapAccountId": "<account-id>",
    "ruleYaml": "name: Invoice triage\ndescription: Invoice triage\nsearch:\n  subject_contains: invoice\noutput:\n  format: json\n  limit: 10\n  fields:\n    - uid\n    - subject\n    - from\nactions:\n  move_to: Archive\n"
  }' | jq
```

```bash
curl -s http://127.0.0.1:8080/api/rules/<rule-id>/dry-run \
  -H 'Content-Type: application/json' \
  -d '{"imapAccountId":"<account-id>"}' | jq
```

The dry-run path:

- validates the stored YAML rule
- resolves the stored account credentials
- runs the existing DSL fetch engine against the selected mailbox
- stores a `rule_runs` record
- returns sample rows and a non-destructive action plan summary

## Automated verification

Unit and integration coverage:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
go test ./pkg/smailnaild/...
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/accounts -run TestServiceAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild/rules -run TestDryRunAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test ./pkg/smailnaild -run TestHostedHTTPFlowAgainstLocalDovecot -v
SMAILNAILD_DOVECOT_TEST=1 go test -tags sqlite_fts5 ./...
```

## Current development assumptions

Until hosted OIDC session handling is implemented:

- all hosted account and rule APIs use a local default user ID
- `X-Smailnail-User-ID` can be supplied to simulate another user during development
- the stored `user_id` column is still the ownership boundary for accounts and rules
