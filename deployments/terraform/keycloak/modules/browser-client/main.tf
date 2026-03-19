terraform {
  required_providers {
    keycloak = {
      source = "keycloak/keycloak"
    }
  }
}

resource "keycloak_openid_client" "this" {
  realm_id                     = var.realm_id
  client_id                    = var.client_id
  name                         = var.name
  enabled                      = var.enabled
  access_type                  = "CONFIDENTIAL"
  client_secret                = var.client_secret
  standard_flow_enabled        = true
  direct_access_grants_enabled = false
  service_accounts_enabled     = false
  use_refresh_tokens           = var.use_refresh_tokens
  valid_redirect_uris          = var.valid_redirect_uris
  web_origins                  = var.web_origins
}

resource "keycloak_openid_client_default_scopes" "this" {
  count          = var.manage_scope_attachments ? 1 : 0
  realm_id       = var.realm_id
  client_id      = keycloak_openid_client.this.id
  default_scopes = var.default_scopes
}

resource "keycloak_openid_client_optional_scopes" "this" {
  count           = var.manage_scope_attachments ? 1 : 0
  realm_id        = var.realm_id
  client_id       = keycloak_openid_client.this.id
  optional_scopes = var.optional_scopes
}
