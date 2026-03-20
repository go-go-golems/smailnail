# smailnail Keycloak Terraform

This directory is no longer the canonical hosted Terraform location.

Use the shared infra repo instead:

- [keycloak/README.md](/home/manuel/code/wesen/terraform/keycloak/README.md)
- [apps/smailnail/envs/local](/home/manuel/code/wesen/terraform/keycloak/apps/smailnail/envs/local/main.tf)
- [apps/smailnail/envs/hosted](/home/manuel/code/wesen/terraform/keycloak/apps/smailnail/envs/hosted/main.tf)

The original repo-local scaffold remains as historical context for how the
first Terraform version was developed before centralization.

This directory is the initial Terraform scaffold for managing the `smailnail`
Keycloak setup declaratively.

It is intentionally split into:

- `modules/realm-base`
- `modules/browser-client`
- `modules/mcp-client`
- `modules/local-fixtures`
- `envs/local`
- `envs/hosted`

The current goal is parity with the existing documented setup:

- local `smailnail-dev` realm imported from JSON today
- hosted `smailnail` realm configured imperatively today
- `smailnail-web` as the browser-login client
- `smailnail-mcp` as the baseline MCP client

This scaffold does not yet attempt to manage every Keycloak policy surface.
Client-registration policy and advanced mapper management should be added only
after the base realm and client lifecycle is stable under Terraform.

## Local sandbox verification

The local environment can authenticate directly to the local Keycloak bootstrap
admin using `admin-cli` and create a sandbox realm without colliding with the
existing imported `smailnail-dev` realm.

From:

```bash
cd /home/manuel/workspaces/2026-03-08/update-imap-mcp/smailnail/deployments/terraform/keycloak/envs/local
```

Run:

```bash
terraform init -backend=false
terraform validate
terraform plan \
  -var='keycloak_url=http://127.0.0.1:18080' \
  -var='realm_name=smailnail-dev-tf' \
  -var='realm_display_name=smailnail-dev-tf' \
  -var='web_client_secret=smailnail-web-secret'
terraform apply -auto-approve \
  -var='keycloak_url=http://127.0.0.1:18080' \
  -var='realm_name=smailnail-dev-tf' \
  -var='realm_display_name=smailnail-dev-tf' \
  -var='web_client_secret=smailnail-web-secret'
```

Verify:

```bash
curl -fsS \
  http://127.0.0.1:18080/realms/smailnail-dev-tf/.well-known/openid-configuration \
  | jq -r '.issuer'
```

Expected output:

```text
http://127.0.0.1:18080/realms/smailnail-dev-tf
```

After apply, rerun the same `terraform plan` command. The expected result is:

```text
No changes. Your infrastructure matches the configuration.
```
