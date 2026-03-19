# smailnail merged hosted server on Coolify

This document describes the merged hosted deployment shape where one `smailnaild`
process serves:

- the hosted web UI and `/api/*`
- browser login routes under `/auth/*`
- MCP metadata at `/.well-known/oauth-protected-resource`
- MCP streamable HTTP at `/mcp`

The merged server keeps the web session flow and the MCP bearer-token flow
separate even though both are mounted on the same `http.Server`.

## Target shape

- Public app URL: `https://smailnail.scapegoat.dev`
- Public MCP URL: `https://smailnail.scapegoat.dev/mcp`
- Keycloak issuer: `https://auth.scapegoat.dev/realms/smailnail`
- Shared app DB: the same DB used for browser sessions, users, IMAP accounts, and rules
- Shared encryption key: the same key used to encrypt stored IMAP credentials for both the web app and MCP

## Runtime routes

The merged runtime serves these route families:

- `/`
  - Vite-built SPA served by `smailnaild`
- `/auth/login`
  - browser redirect to Keycloak
- `/auth/callback`
  - OIDC callback that provisions or refreshes the local user and creates the session cookie
- `/auth/logout`
  - local session logout
- `/api/*`
  - session-backed web API
- `/.well-known/oauth-protected-resource`
  - public MCP/OAuth metadata
- `/mcp`
  - bearer-token-protected MCP endpoint

That means the public MCP resource URL advertised to clients must be:

```text
https://smailnail.scapegoat.dev/mcp
```

The `/.well-known/oauth-protected-resource` endpoint is mounted at the host root,
not under `/mcp`.

## Container build

The root [Dockerfile](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile) now builds the merged hosted server:

- stage 1 builds the Vite UI
- stage 2 builds `cmd/smailnaild` with `-tags embed`
- stage 3 packages the single hosted binary plus the merged entrypoint

The container entrypoint is
[docker-entrypoint.smailnaild.sh](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/scripts/docker-entrypoint.smailnaild.sh).

If started without explicit arguments, the image runs:

```text
smailnaild serve
```

with config assembled from environment variables.

## Required environment variables

### Shared app runtime

- `SMAILNAILD_LISTEN_PORT`
- `SMAILNAILD_DB_TYPE`
- `SMAILNAILD_DATABASE` or `SMAILNAILD_DSN`
- `SMAILNAILD_ENCRYPTION_KEY_ID`
- `SMAILNAILD_ENCRYPTION_KEY_BASE64`
- `SMAILNAILD_LOG_LEVEL`

### Web OIDC

- `SMAILNAILD_AUTH_MODE=oidc`
- `SMAILNAILD_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`
- `SMAILNAILD_OIDC_CLIENT_ID=smailnail-web`
- `SMAILNAILD_OIDC_CLIENT_SECRET=...`
- `SMAILNAILD_OIDC_REDIRECT_URL=https://smailnail.scapegoat.dev/auth/callback`
- `SMAILNAILD_OIDC_SCOPES=openid,profile,email`

### MCP OIDC

- `SMAILNAILD_MCP_ENABLED=1`
- `SMAILNAILD_MCP_TRANSPORT=streamable_http`
- `SMAILNAILD_MCP_AUTH_MODE=external_oidc`
- `SMAILNAILD_MCP_AUTH_RESOURCE_URL=https://smailnail.scapegoat.dev/mcp`
- `SMAILNAILD_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail`

Optional MCP tightening once the Keycloak client is ready:

- `SMAILNAILD_MCP_OIDC_AUDIENCE=smailnail-mcp`
- `SMAILNAILD_MCP_OIDC_REQUIRED_SCOPES=mcp:invoke`

## Recommended Coolify environment

```env
SMAILNAILD_LISTEN_PORT=8080
SMAILNAILD_DB_TYPE=sqlite
SMAILNAILD_DATABASE=/data/smailnaild.sqlite
SMAILNAILD_LOG_LEVEL=debug
SMAILNAILD_ENCRYPTION_KEY_ID=prod-smailnail
SMAILNAILD_ENCRYPTION_KEY_BASE64=...

SMAILNAILD_AUTH_MODE=oidc
SMAILNAILD_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail
SMAILNAILD_OIDC_CLIENT_ID=smailnail-web
SMAILNAILD_OIDC_CLIENT_SECRET=...
SMAILNAILD_OIDC_REDIRECT_URL=https://smailnail.scapegoat.dev/auth/callback
SMAILNAILD_OIDC_SCOPES=openid,profile,email

SMAILNAILD_MCP_ENABLED=1
SMAILNAILD_MCP_TRANSPORT=streamable_http
SMAILNAILD_MCP_AUTH_MODE=external_oidc
SMAILNAILD_MCP_AUTH_RESOURCE_URL=https://smailnail.scapegoat.dev/mcp
SMAILNAILD_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail
```

## Coolify application shape

- Build pack: Dockerfile
- Dockerfile path: `Dockerfile`
- Exposed port: `8080`
- Domain: `https://smailnail.scapegoat.dev`
- Health check path: `/readyz`

The health check should stay on the web app side rather than `/mcp`, because:

- `/readyz` is a normal app liveness/readiness signal
- `/mcp` intentionally returns `401` without a bearer token
- `/.well-known/oauth-protected-resource` is public, but it only proves the MCP auth metadata path, not the whole web app

## Local build and container smoke

Build the merged image:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker build -t smailnaild-merged:dev .
```

Run the merged image locally against the local Keycloak fixture:

```bash
docker run --rm -p 8081:8080 \
  -e SMAILNAILD_DB_TYPE=sqlite \
  -e SMAILNAILD_DATABASE=/tmp/smailnaild.sqlite \
  -e SMAILNAILD_ENCRYPTION_KEY_ID="$SMAILNAILD_ENCRYPTION_KEY_ID" \
  -e SMAILNAILD_ENCRYPTION_KEY_BASE64="$SMAILNAILD_ENCRYPTION_KEY_BASE64" \
  -e SMAILNAILD_AUTH_MODE=oidc \
  -e SMAILNAILD_OIDC_ISSUER_URL=http://host.docker.internal:18080/realms/smailnail-dev \
  -e SMAILNAILD_OIDC_CLIENT_ID=smailnail-web \
  -e SMAILNAILD_OIDC_CLIENT_SECRET=smailnail-web-secret \
  -e SMAILNAILD_OIDC_REDIRECT_URL=http://127.0.0.1:8081/auth/callback \
  -e SMAILNAILD_MCP_ENABLED=1 \
  -e SMAILNAILD_MCP_AUTH_MODE=external_oidc \
  -e SMAILNAILD_MCP_AUTH_RESOURCE_URL=http://127.0.0.1:8081/mcp \
  -e SMAILNAILD_MCP_OIDC_ISSUER_URL=http://host.docker.internal:18080/realms/smailnail-dev \
  smailnaild-merged:dev
```

Then verify:

```bash
curl -s http://127.0.0.1:8081/readyz | jq
curl -s http://127.0.0.1:8081/.well-known/oauth-protected-resource | jq
curl -i -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  http://127.0.0.1:8081/mcp
```

## Hosted verification order

After deployment, validate in this order:

1. `GET /readyz`
2. `GET /.well-known/oauth-protected-resource`
3. unauthenticated `POST /mcp` returns `401`
4. browser login through `/auth/login`
5. `GET /api/me` returns the session-backed local user
6. create an IMAP account through the UI or `/api/accounts`
7. run the account test against the hosted Dovecot fixture
8. obtain an OIDC token and call `/mcp` using the stored `accountId`

## Standalone MCP compatibility

The standalone
[smailnail-imap-mcp main.go](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/cmd/smailnail-imap-mcp/main.go)
still exists as a compatibility wrapper for:

- focused local MCP debugging
- old hosted rollback if the merged deployment fails
- tool-specific testing without the web app

The root production image is now the merged hosted server, not the standalone MCP binary.

## Cutover notes

The old public MCP host is `https://smailnail.mcp.scapegoat.dev/mcp`.

After the merged server is validated on `https://smailnail.scapegoat.dev/mcp`, the old app can either:

- remain as a temporary rollback target
- or be retired once clients are moved to the merged host

Do not cut over clients until both of these are true:

- browser login and account setup are working on `smailnail.scapegoat.dev`
- bearer-authenticated `executeIMAPJS` works against `smailnail.scapegoat.dev/mcp`
