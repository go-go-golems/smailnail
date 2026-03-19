---
name: smailnail-coolify-deploy
description: Redeploy smailnail to Coolify, including the go-go-mcp dependency handoff, remote Coolify CLI usage over SSH, and post-deploy MCP verification. Use when asked to deploy or redeploy smailnail, the hosted smailnail MCP, or the merged smailnaild server on Coolify.
---

# Smailnail Coolify Deploy

Use this skill when a change in this repo needs to be live on Coolify.

## Read first

- For the merged hosted server, read `docs/deployments/smailnaild-merged-coolify.md`.
- For the legacy MCP-only host, read `docs/deployments/smailnail-imap-mcp-coolify.md`.

## Deployment workflow

1. If the change touches `go-go-mcp`, publish that repo first.
   Run `go test ./pkg/embeddable` in the `go-go-mcp` repo, commit the MCP change, and push it.
2. Bump `smailnail` to the exact `go-go-mcp` commit.
   Run `go get github.com/go-go-golems/go-go-mcp@<commit>` from the `smailnail` repo so `go.mod` and `go.sum` point at the deployed library revision.
3. Validate the app repo before shipping.
   Run `go test ./pkg/smailnaild ./pkg/mcp/imapjs`.
4. Commit and push the `smailnail` branch that Coolify builds.
   Do not rely on local workspace state; Coolify only sees pushed commits.
5. Redeploy from the Coolify host over SSH.
   This workspace does not have a local `scapegoat` Coolify context configured, so use the server-side CLI.

## Remote Coolify commands

Default server:

```bash
ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify context verify --context scapegoat
  coolify app list --context scapegoat --format json
'
```

If you already know the app UUID, deploy it directly:

```bash
ssh root@89.167.52.236 '
  export PATH=$PATH:/usr/local/bin:$HOME/go/bin
  coolify deploy uuid <app-uuid> --context scapegoat --force --format pretty
  coolify app get <app-uuid> --context scapegoat --format pretty
'
```

Known legacy MCP app UUID from the current deployment notes:

- `fhp3mxqlfftdxdib3vxz89l3` for `smailnail.mcp.scapegoat.dev`

## Post-deploy verification

Run these checks against the public host:

```bash
curl -fsS https://<host>/.well-known/oauth-protected-resource | jq
curl -i -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}' \
  https://<host>/mcp
curl -fsS https://<host>/readyz | jq
```

Expected:

- `/.well-known/oauth-protected-resource` returns the exact public `/mcp` URL in `resource`
- unauthenticated `POST /mcp` returns `401`
- `/readyz` returns `200`

For Claude OAuth debugging, inspect app logs after redeploy:

- the `WWW-Authenticate` challenge should only advertise `resource_metadata`
- Claude should progress beyond `/.well-known/oauth-protected-resource` to issuer metadata or `/authorize`
- repeated `401` responses on `/mcp` without any auth-server discovery usually indicate a client-side OAuth handoff problem, not a broken bearer validator
