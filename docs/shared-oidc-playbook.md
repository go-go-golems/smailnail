# Shared OIDC Playbook

This playbook shows how to exercise the shared identity flow across:

- `smailnaild` browser login
- the shared application database
- stored IMAP account ownership
- merged `/mcp` bearer-authenticated execution

The shared-identity contract is:

- web login resolves a local user from `(issuer, subject)`
- MCP bearer auth resolves the same local user from `(issuer, subject)`
- IMAP accounts belong to `users.id`
- MCP account access is allowed only when the stored account belongs to that same local user

## 1. Start the local fixtures

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
docker compose -f docker-compose.local.yml up -d dovecot keycloak-postgres keycloak
```

Local defaults:

- Keycloak issuer: `http://127.0.0.1:18080/realms/smailnail-dev`
- Keycloak admin: `admin` / `admin`
- Dovecot IMAPS: `127.0.0.1:993`
- Dovecot test user: `a`
- Dovecot password: `pass`

## 2. Start the merged `smailnaild` with OIDC and MCP enabled

This keeps the hosted account database local and reusable by both the browser and MCP.

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail

go run ./cmd/smailnaild serve \
  --listen-port 8080 \
  --db-type sqlite \
  --database ./smailnaild.sqlite \
  --encryption-key-id "$SMAILNAILD_ENCRYPTION_KEY_ID" \
  --encryption-key-base64 "$SMAILNAILD_ENCRYPTION_KEY_BASE64" \
  --auth-mode oidc \
  --oidc-issuer-url http://127.0.0.1:18080/realms/smailnail-dev \
  --oidc-client-id smailnail-web \
  --oidc-client-secret smailnail-web-secret \
  --oidc-redirect-url http://127.0.0.1:8080/auth/callback \
  --mcp-enabled \
  --mcp-auth-mode external_oidc \
  --mcp-auth-resource-url http://127.0.0.1:8080/mcp \
  --mcp-oidc-issuer-url http://127.0.0.1:18080/realms/smailnail-dev
```

## 3. Log in through the browser

Today the frontend does not yet auto-redirect, so use the backend route directly:

- `http://localhost:8080/auth/login`

After login, verify the session:

```bash
curl -i \
  --cookie-jar /tmp/smailnail.cookies \
  --cookie /tmp/smailnail.cookies \
  http://localhost:8080/api/me
```

Expected outcome:

- `200 OK`
- JSON payload with:
  - local `id`
  - `issuer`
  - `subject`
  - optional profile fields such as `email` and `displayName`

## 4. Add a stored IMAP account

Use the UI or the hosted API to create an account owned by the logged-in user.

Current local test Dovecot settings:

- server: `127.0.0.1`
- port: `993`
- username: `a`
- password: `pass`
- mailbox default: `INBOX`
- insecure TLS: `true`

Minimal API example:

```bash
curl -s \
  --cookie /tmp/smailnail.cookies \
  -H 'Content-Type: application/json' \
  -d '{
    "label": "Local Dovecot",
    "server": "127.0.0.1",
    "port": 993,
    "username": "a",
    "password": "pass",
    "mailboxDefault": "INBOX",
    "insecure": true,
    "authKind": "password",
    "mcpEnabled": true
  }' \
  http://localhost:8080/api/accounts | jq
```

Record the returned `id`.

## 5. Use the merged `/mcp` endpoint on the same server

The crucial part is that the same `smailnaild` process now owns both the browser
session flow and the MCP bearer-token flow. There is no second server process to
start for the normal hosted path.

## 6. Fetch an access token from Keycloak

The local integration test uses password grant for convenience. That is acceptable for local test automation only.

```bash
TOKEN=$(
  curl -s \
    -d grant_type=password \
    -d client_id=smailnail-mcp \
    -d username=alice \
    -d password=secret \
    http://127.0.0.1:18080/realms/smailnail-dev/protocol/openid-connect/token \
  | jq -r .access_token
)
```

## 7. Call the hosted MCP with a stored account

Use the `accountId` path added to the JavaScript service:

```bash
curl -s \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": "1",
    "method": "tools/call",
    "params": {
      "name": "executeIMAPJS",
      "arguments": {
        "code": "const smailnail = require(\"smailnail\"); const svc = smailnail.newService(); const session = svc.connect({ accountId: \"ACCOUNT_ID_HERE\" }); const result = { mailbox: session.mailbox }; session.close(); result;"
      }
    }
  }' \
  http://127.0.0.1:8080/mcp | jq
```

Expected outcome:

- success response
- returned payload includes `mailbox: "INBOX"`

This proves:

- the bearer token was validated
- the token identity resolved to the same local user as the browser login
- the stored account belonged to that same local user
- the IMAP password was decrypted from the shared app DB

## 8. Local end-to-end regression test

The repo now includes one command that exercises the same flow with live local Keycloak and live local Dovecot:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail
SMAILNAIL_LOCAL_STACK_TEST=1 go test ./pkg/mcp/imapjs -run TestExecuteIMAPJSAgainstLocalKeycloakAndDovecot -v
```

That test:

- ensures the local Keycloak user and client are usable
- fetches a real local OIDC access token
- provisions the same local user in the app DB
- creates a stored IMAP account
- runs `executeIMAPJS` using `accountId`
- verifies the runtime connects to local Dovecot

## 9. Remote production shape

For production, keep the same identity model and change only the endpoints:

- `smailnaild`
  - issuer: `https://auth.scapegoat.dev/realms/smailnail`
  - redirect URL: `https://smailnail.scapegoat.dev/auth/callback`
- merged MCP surface
  - resource URL: `https://smailnail.scapegoat.dev/mcp`
  - protected resource metadata: `https://smailnail.scapegoat.dev/.well-known/oauth-protected-resource`
  - issuer: `https://auth.scapegoat.dev/realms/smailnail`
- shared application DB:
  - production SQLite on one host or PostgreSQL for shared deployment

The production difference is transport and hosting, not identity semantics.

## 10. Failure patterns

If browser login works but MCP cannot use a stored account, check these first:

- `smailnail-imap-mcp` is pointing at the same application DB as `smailnaild`
- the MCP process has the same encryption key ID and base64 key
- the stored account has `mcpEnabled: true`
- the MCP access token `iss` and `sub` match the web-login identity

If connector-based OAuth fails before token issuance, inspect Keycloak client-registration policies separately. That is a provider-side problem, not a shared-user-mapping problem.
