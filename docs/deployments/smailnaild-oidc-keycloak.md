# smailnaild OIDC with Keycloak

This document describes the hosted web-login setup for `smailnaild`.

For the merged production deployment where the same binary also serves `/mcp`,
see
[smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md).

The implementation is server-side OIDC:

- browser redirects to the OIDC provider
- `smailnaild` exchanges the authorization code
- `smailnaild` verifies the `id_token`
- `smailnaild` provisions or refreshes the local user using `(issuer, subject)`
- `smailnaild` stores a local session cookie

This is intentionally separate from MCP bearer auth, but both flows resolve into the same local user tables.

## Current implementation constraints

- `smailnaild` currently implements authorization code flow with state and nonce cookies.
- It does not currently implement PKCE.
- For that reason, the simplest supported Keycloak setup today is a confidential client with a client secret.
- The frontend is not yet auto-redirecting anonymous users to `/auth/login`, so login is easiest to test by visiting `/auth/login` directly.

## Keycloak client shape

Recommended web client:

- Realm: `smailnail` in production, `smailnail-dev` locally
- Client ID: `smailnail-web`
- Client type: OpenID Connect
- Access type: confidential
- Standard flow: enabled
- Direct access grants: disabled
- Service accounts: disabled

Recommended redirect URIs:

- Local backend:
  - `http://localhost:8080/auth/callback`
  - `http://localhost:3001/auth/callback`
- Hosted backend:
  - `https://smailnail.scapegoat.dev/auth/callback`

Recommended web origins:

- `http://localhost:5050`
- `http://localhost:8080`
- `http://localhost:3001`
- `https://smailnail.scapegoat.dev`

Recommended protocol scopes:

- `openid`
- `profile`
- `email`

Useful optional mappers:

- `email`
- `email_verified`
- `preferred_username`
- `name`
- `picture`

Those claims are treated as profile metadata only. The stable local identity key remains `(issuer, subject)`.

## `smailnaild` command settings

The hosted backend reads OIDC settings from the `auth` Glazed section defined in:

- `pkg/smailnaild/auth/config.go`

Relevant flags:

- `--auth-mode oidc`
- `--auth-session-cookie-name smailnail_session`
- `--oidc-issuer-url ...`
- `--oidc-client-id ...`
- `--oidc-client-secret ...`
- `--oidc-redirect-url ...`
- `--oidc-scopes openid,profile,email`

It also needs:

- Clay SQL connection settings for the shared app database
- encryption key settings for IMAP credential storage

## Local example

This local example assumes:

- Keycloak is running from `docker-compose.local.yml`
- the imported realm is `smailnail-dev`
- the local backend listens on `8080`

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail

go run ./cmd/smailnaild serve \
  --listen-port 8080 \
  --db-type sqlite \
  --database ./smailnaild.sqlite \
  --encryption-key-id "$SMAILNAILD_ENCRYPTION_KEY_ID" \
  --encryption-key-base64 "$SMAILNAILD_ENCRYPTION_KEY_BASE64" \
  --auth-mode oidc \
  --auth-session-cookie-name smailnail_session \
  --oidc-issuer-url http://127.0.0.1:18080/realms/smailnail-dev \
  --oidc-client-id smailnail-web \
  --oidc-client-secret smailnail-web-secret \
  --oidc-redirect-url http://127.0.0.1:8080/auth/callback
```

Then open:

- `http://localhost:8080/auth/login`

After login, verify:

```bash
curl -i --cookie-jar /tmp/smailnail.cookies --cookie /tmp/smailnail.cookies \
  http://localhost:8080/api/me
```

If you want the Vite UI in front of the backend:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/ui
SMAILNAIL_UI_BACKEND_PORT=8080 pnpm run dev
```

Use:

- `http://localhost:5050`

The current frontend slice is not yet redirecting through `/auth/login`, so the browser login still needs to happen explicitly first.

## Production example

```bash
smailnaild serve \
  --db-type postgres \
  --dsn "$SMAILNAILD_DATABASE_DSN" \
  --encryption-key-id "$SMAILNAILD_ENCRYPTION_KEY_ID" \
  --encryption-key-base64 "$SMAILNAILD_ENCRYPTION_KEY_BASE64" \
  --auth-mode oidc \
  --auth-session-cookie-name smailnail_session \
  --oidc-issuer-url https://auth.scapegoat.dev/realms/smailnail \
  --oidc-client-id smailnail-web \
  --oidc-client-secret "$SMAILNAILD_OIDC_CLIENT_SECRET" \
  --oidc-redirect-url https://smailnail.scapegoat.dev/auth/callback
```

For production, put `smailnaild` behind HTTPS and set the Keycloak client redirect URIs and web origins to the public hostnames only.

If you are deploying the merged hosted server, add the MCP flags described in
[smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md)
so the same process also serves `/.well-known/oauth-protected-resource` and `/mcp`.

## Identity model

The session created by `smailnaild` is local application state, not a copy of the whole token set.

The important identity flow is:

```text
Keycloak id_token -> iss + sub -> user_external_identities -> users.id
                                                |
                                                +-> imap_accounts.user_id
                                                +-> rules.user_id
```

That design is intentionally not Keycloak-specific. Any OIDC provider that gives a stable `iss` and `sub` can fit the same local user model.
