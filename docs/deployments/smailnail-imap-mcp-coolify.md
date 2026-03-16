# smailnail-imap-mcp Coolify Deployment

This document describes the first production deployment slice for `smailnail-imap-mcp`.

## Target

- Public MCP URL: `https://smailnail.mcp.scapegoat.dev/mcp`
- Keycloak issuer: `https://auth.scapegoat.dev/realms/smailnail`
- Primary transport: `streamable_http`

## Files

- Default Docker image: `Dockerfile`
- Alternate Docker image: `Dockerfile.smailnail-imap-mcp`
- Container entrypoint: `scripts/docker-entrypoint.smailnail-imap-mcp.sh`

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
- `SMAILNAIL_MCP_EXTRA_ARGS`

## Recommended Coolify environment

```env
SMAILNAIL_MCP_TRANSPORT=streamable_http
SMAILNAIL_MCP_PORT=3201
SMAILNAIL_MCP_AUTH_MODE=external_oidc
SMAILNAIL_MCP_AUTH_RESOURCE_URL=https://smailnail.mcp.scapegoat.dev/mcp
SMAILNAIL_MCP_OIDC_ISSUER_URL=https://auth.scapegoat.dev/realms/smailnail
```

Only set the following once the Keycloak client/realm side is ready:

```env
SMAILNAIL_MCP_OIDC_AUDIENCE=smailnail-mcp
SMAILNAIL_MCP_OIDC_REQUIRED_SCOPES=mcp:invoke
```

## Coolify application shape

- Build pack: Dockerfile
- Dockerfile path: `Dockerfile`
- Exposed port: `3201`
- Domain: `smailnail.mcp.scapegoat.dev`
- Health check path: `/.well-known/oauth-protected-resource`

The health check path is intentionally public when auth is enabled, unlike `/mcp`.
The runtime image must include `curl` or `wget`, because Coolify runs the health check from inside the container.

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
  --domains smailnail.mcp.scapegoat.dev \
  --health-check-enabled \
  --health-check-path /.well-known/oauth-protected-resource
```

Once the branch is pushed, this avoids any registry credentials entirely.

## Keycloak expectations

The production deployment expects a non-`master` realm:

- realm: `smailnail`
- issuer: `https://auth.scapegoat.dev/realms/smailnail`

The MCP client entry in Keycloak must match the actual OAuth consumer. For Claude-based remote MCP, configure the client with the relevant Claude callback URIs and whatever audience/scope mapping you decide to enforce.

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
