# smailnail Dovecot Fixture on Coolify

This document describes the hosted IMAP fixture used to test `smailnail` and `smailnail-imap-mcp` against a real remote mail server.

## Goal

Mirror the local `docker-test-dovecot` setup on the Hetzner/Coolify host so remote tests can use the same predictable test users and self-signed TLS behavior.

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
go run ./cmd/smailnail fetch-mail \
  --server 89.167.52.236 \
  --port 993 \
  --username a \
  --password pass \
  --insecure \
  --output yaml
```

Then mailbox/action validation can use the same `imap-tests` and `smailnail` commands as the local smoke script, just pointed at the hosted IP.
