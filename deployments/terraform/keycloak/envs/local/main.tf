module "realm" {
  source       = "../../modules/realm-base"
  realm_name   = var.realm_name
  display_name = var.realm_display_name
}

module "browser_client" {
  source        = "../../modules/browser-client"
  realm_id      = module.realm.id
  client_id     = "smailnail-web"
  name          = "smailnail-web"
  client_secret = var.web_client_secret
  valid_redirect_uris = [
    "http://localhost:5050/*",
    "http://127.0.0.1:5050/*",
    "http://localhost:8080/*",
    "http://127.0.0.1:8080/*",
    "http://localhost:8081/*",
    "http://127.0.0.1:8081/*",
  ]
  web_origins = [
    "http://localhost:5050",
    "http://127.0.0.1:5050",
    "http://localhost:8080",
    "http://127.0.0.1:8080",
    "http://localhost:8081",
    "http://127.0.0.1:8081",
  ]
}

module "mcp_client" {
  source      = "../../modules/mcp-client"
  realm_id    = module.realm.id
  client_id   = "smailnail-mcp"
  name        = "smailnail-mcp"
  access_type = "PUBLIC"
  valid_redirect_uris = [
    "http://localhost/*",
    "http://127.0.0.1/*",
  ]
  web_origins = ["+"]
}

module "local_fixtures" {
  source         = "../../modules/local-fixtures"
  realm_id       = module.realm.id
  alice_password = var.alice_password
}
