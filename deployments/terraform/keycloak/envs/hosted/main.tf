locals {
  anonymous_dcr_allowed_client_scopes = [
    "mcp:tools",
    "openid",
    "service_account",
    "web-origins",
  ]
}

module "realm" {
  source                      = "../../modules/realm-base"
  realm_name                  = "smailnail"
  display_name                = "smailnail"
  default_signature_algorithm = "RS256"
}

module "browser_client" {
  source                   = "../../modules/browser-client"
  realm_id                 = module.realm.id
  client_id                = "smailnail-web"
  name                     = null
  client_secret            = var.web_client_secret
  use_refresh_tokens       = false
  manage_scope_attachments = false
  valid_redirect_uris = [
    "https://smailnail.mcp.scapegoat.dev/auth/callback",
  ]
  web_origins = [
    "https://smailnail.mcp.scapegoat.dev",
  ]
}

module "mcp_client" {
  source                       = "../../modules/mcp-client"
  realm_id                     = module.realm.id
  client_id                    = "smailnail-mcp"
  name                         = "smailnail-mcp"
  access_type                  = "PUBLIC"
  client_secret                = var.mcp_client_secret
  direct_access_grants_enabled = false
  use_refresh_tokens           = false
  manage_scope_attachments     = false
  valid_redirect_uris = [
    "https://claude.ai/api/mcp/auth_callback",
    "https://claude.com/api/mcp/auth_callback",
    "https://smailnail.mcp.scapegoat.dev/*",
  ]
  web_origins = ["+"]
}

resource "terraform_data" "anonymous_dcr_allowed_client_scopes" {
  triggers_replace = [
    var.keycloak_url,
    var.keycloak_admin_realm,
    module.realm.realm,
    jsonencode(local.anonymous_dcr_allowed_client_scopes),
  ]

  provisioner "local-exec" {
    command = "${path.module}/../../scripts/update_anonymous_dcr_allowed_client_scopes.sh"

    environment = {
      KEYCLOAK_URL         = var.keycloak_url
      KEYCLOAK_ADMIN_REALM = var.keycloak_admin_realm
      KEYCLOAK_CLIENT_ID   = var.keycloak_client_id
      KEYCLOAK_USERNAME    = var.keycloak_username
      KEYCLOAK_PASSWORD    = var.keycloak_password
      REALM_NAME           = module.realm.realm
      DESIRED_SCOPES_JSON  = jsonencode(local.anonymous_dcr_allowed_client_scopes)
    }
  }

  depends_on = [
    module.realm,
    module.mcp_client,
  ]
}
