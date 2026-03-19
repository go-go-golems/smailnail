module "realm" {
  source       = "../../modules/realm-base"
  realm_name   = "smailnail"
  display_name = "smailnail"
}

module "browser_client" {
  source        = "../../modules/browser-client"
  realm_id      = module.realm.id
  client_id     = "smailnail-web"
  name          = "smailnail-web"
  client_secret = var.web_client_secret
  valid_redirect_uris = [
    "https://smailnail.scapegoat.dev/auth/callback",
  ]
  web_origins = [
    "https://smailnail.scapegoat.dev",
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
  valid_redirect_uris = [
    "https://claude.ai/api/mcp/auth_callback",
    "https://claude.com/api/mcp/auth_callback",
    "https://smailnail.scapegoat.dev/*",
  ]
  web_origins = ["+"]
}
