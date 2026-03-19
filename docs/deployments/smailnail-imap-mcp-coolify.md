# smailnail-imap-mcp Coolify Deployment

This document now describes the legacy standalone MCP deployment shape.

The merged hosted deployment is documented in
[smailnaild-merged-coolify.md](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/docs/deployments/smailnaild-merged-coolify.md)
and is the preferred production target going forward.

This document describes the first production deployment slice for `smailnail-imap-mcp`.

## Target

Legacy standalone target:

- Public MCP URL: `https://smailnail.mcp.scapegoat.dev/mcp`
- Keycloak issuer: `https://auth.scapegoat.dev/realms/smailnail`
- Primary transport: `streamable_http`

## Files

- Default Docker image: `Dockerfile.smailnail-imap-mcp`
- Alternate Docker image: `Dockerfile.smailnail-imap-mcp`
- Container entrypoint: `scripts/docker-entrypoint.smailnail-imap-mcp.sh`

The repository root [Dockerfile](/home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/Dockerfile)
now builds the merged hosted server instead.

## Container defaults

If the container is started without arguments, it runs:

```text
smailnail-imap-mcp mcp start --transport streamable_http --port 3201
```

The entrypoint can be configured entirely through environment variables:

- `SMAILNAIL_MCP_TRANSPORT`
- `SMAILNAIL_MCP_PORT`
- `SMAILNAIL_MCP_AUTH_MODE`
- `SMAILNAIL_MCP_AUTH_RESOURCE_URL`
- `SMAILNAIL_MCP_OIDC_ISSUER_URL`
- `SMAILNAIL_MCP_OIDC_DISCOVERY_URL`
- `SMAILNAIL_MCP_OIDC_AUDIENCE`
- `SMAILNAIL_MCP_OIDC_REQUIRED_SCOPES`
- `SMAILNAIL_MCP_APP_DB_DRIVER`
- `SMAILNAIL_MCP_APP_DB_DSN`
- `SMAILNAIL_MCP_APP_ENCRYPTION_KEY_ID`
- `SMAILNAIL_MCP_APP_ENCRYPTION_KEY_BASE64`
- `SMAILNAIL_MCP_EXTRA_ARGS`

## Recommended Coolify environment

```env
SMAILNAIL_MCP_TRANSPORT=streamable_http
SMAILNAIL_MCP_PORT=3201
SMAILNAIL_MCP_AUTH_MODE=external_oidc
SMAILNAIL_MCP_AUTH_RESOURCE_URL=https://smailnail.mcp.scapegoat.dev/mcp
SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail
SMAILNAIL_MCP_APP_DB_DRIVER=pgx
SMAILNAIL_MCP_APP_DB_DSN=postgres://smailnail:...@postgres:5432/smailnail?sslmode=disable
SMAILNAIL_MCP_APP_ENCRYPTION_KEY_ID=prod-smailnail
SMAILNAIL_MCP_APP_ENCRYPTION_KEY_BASE64=...
```

Only set the following once the Keycloak client/realm side is ready:

```env
SMAILNAIL_MCP_OIDC_AUDIENCE=smailnail-mcp
SMAILNAIL_MCP_OIDC_REQUIRED_SCOPES=mcp:invoke
```

The shared app DB and encryption settings are required for the new identity-to-account path. Without them:

- bearer auth can still validate successfully
- the MCP can still identify the user
- but stored IMAP accounts created by `smailnaild` cannot be resolved or decrypted

## Coolify application shape

- Build pack: Dockerfile
- Dockerfile path: `Dockerfile`
- Exposed port: `3201`
- Domain: `smailnail.mcp.scapegoat.dev`
- Health check path: `/.well-known/oauth-protected-resource`

The health check path is intentionally public when auth is enabled, unlike `/mcp`.
The runtime image must include `curl` or `wget`, because Coolify runs the health check from inside the container.

## How `/mcp` is routed

Coolify routes the full host `https://smailnail.mcp.scapegoat.dev` to container port `3201`.
There is no Coolify-side path rewrite for the MCP endpoint.

That means:

- `https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource` goes directly to the binary's public metadata handler
- `https://smailnail.mcp.scapegoat.dev/mcp` goes directly to the binary's MCP HTTP handler

The `/mcp` path comes from `smailnail-imap-mcp` itself, not from Traefik or a reverse-proxy rewrite rule that we wrote by hand.

## Public repo deployment

This repository is public, which makes the deployment path simpler than a private registry or private repo flow. The intended Coolify create command shape is:

```bash
coolify app create public \
  --server-uuid cgl105090ljoxitdf7gmvbrm \
  --project-uuid n8xkgqpbjj04m4pishy3su5e \
  --environment-name production \
  --name smailnail-imap-mcp \
  --git-repository https://github.com/wesen/smailnail \
  --git-branch task/update-imap-mcp \
  --build-pack dockerfile \
  --ports-exposes 3201 \
  --domains https://smailnail.mcp.scapegoat.dev \
  --health-check-enabled \
  --health-check-path /.well-known/oauth-protected-resource
```

Once the branch is pushed, this avoids any registry credentials entirely.

## Keycloak expectations

The production deployment expects a non-`master` realm:

- realm: `smailnail`
- issuer: `https://auth.scapegoat.dev/realms/smailnail`

The MCP client entry in Keycloak must match the actual OAuth consumer. For Claude-based remote MCP, configure the client with the relevant Claude callback URIs and whatever audience/scope mapping you decide to enforce.

This deployment also assumes `smailnaild` and `smailnail-imap-mcp` share the same application database so both the web app and the MCP resolve the same local `users.id` and the same `imap_accounts` rows.

## Verification

Unauthenticated metadata:

```bash
curl -s https://smailnail.mcp.scapegoat.dev/.well-known/oauth-protected-resource | jq
```

Unauthenticated MCP should fail with `401` and `WWW-Authenticate`:

```bash
curl -i \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  https://smailnail.mcp.scapegoat.dev/mcp
```

## Local image build

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker build -t smailnail-imap-mcp:dev .
```

## Local container smoke

```bash
docker run --rm -p 3201:3201 \
  -e SMAILNAIL_MCP_AUTH_MODE=external_oidc \
  -e SMAILNAIL_MCP_AUTH_RESOURCE_URL=http://127.0.0.1:3201/mcp \
  -e SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/master \
  smailnail-imap-mcp:dev
```

Then in another shell:

```bash
curl -s http://127.0.0.1:3201/.well-known/oauth-protected-resource | jq
curl -i \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  http://127.0.0.1:3201/mcp
```
