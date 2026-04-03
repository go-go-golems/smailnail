# smailnail Dovecot Fixture on Coolify

This document describes the hosted IMAP fixture used to test `smailnail` and `smailnail-imap-mcp` against a real remote mail server.

## Goal

Mirror the local `docker-test-dovecot` setup on the Hetzner/Coolify host so remote tests can use the same predictable test users and self-signed TLS behavior.

Current hosted service:

- Coolify service UUID: `gh32795yh1av2dpi2j6lhn6h`
- Coolify service name: `smailnail-dovecot-fixture`
- Current test host: `89.167.52.236`

## Compose source

- Compose file: `deployments/coolify/smailnail-dovecot.compose.yaml`
- Image: `ghcr.io/spezifisch/docker-test-dovecot:latest`

## Runtime shape

The hosted fixture exposes the same raw mail ports as the local setup:

- `24` LMTP
- `110` POP3
- `143` IMAP
- `993` IMAPS
- `995` POP3S
- `4190` ManageSieve

There is no HTTP routing layer for this service. These are direct host port bindings on the Coolify machine.
Because this is a raw-port service, Coolify currently reports it as `running:unknown` rather than `running:healthy`; no HTTP-style health check is configured.

## Persistence

The service persists:

- `/home` for user Maildirs
- `/etc/dovecot/ssl` for the generated self-signed TLS material

## Test users

The upstream fixture creates these users with password `pass`:

- `a`
- `b`
- `c`
- `d`
- `rxa`
- `rxb`
- `rxc`
- `rxd`

## Testing expectations

Because the fixture uses self-signed TLS, remote IMAPS checks should use `--insecure`.

Typical validation flow:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail fetch-mail \
  --server 89.167.52.236 \
  --port 993 \
  --username a \
  --password pass \
  --insecure \
  --output yaml
```

Then mailbox/action validation can use the same `imap-tests` and `smailnail` commands as the local smoke script, just pointed at the hosted IP.

## Deterministic hosted validation

Create a mailbox:

```bash
go run ./cmd/imap-tests create-mailbox \
  --server 89.167.52.236 \
  --port 993 \
  --username a \
  --password pass \
  --mailbox INBOX \
  --new-mailbox Archive \
  --insecure \
  --output json
```

Store a known message:

```bash
go run ./cmd/imap-tests store-text-message \
  --server 89.167.52.236 \
  --port 993 \
  --username a \
  --password pass \
  --mailbox INBOX \
  --from 'Remote Seeder <seed@example.com>' \
  --to 'User A <a@testcot>' \
  --subject 'Hosted Coolify Dovecot Test' \
  --body 'Remote hosted IMAP fixture validation.' \
  --insecure \
  --output json
```

Fetch it back:

```bash
go run -tags sqlite_fts5 ./cmd/smailnail fetch-mail \
  --server 89.167.52.236 \
  --port 993 \
  --username a \
  --password pass \
  --mailbox INBOX \
  --subject-contains 'Hosted Coolify Dovecot Test' \
  --insecure \
  --output json
```

Expected result:

- `messages_fetched=1`
- a message with subject `Hosted Coolify Dovecot Test`
- content `Remote hosted IMAP fixture validation.`
