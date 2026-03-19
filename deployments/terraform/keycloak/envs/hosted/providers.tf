provider "keycloak" {
  url           = var.keycloak_url
  base_path     = var.keycloak_base_path
  realm         = var.keycloak_admin_realm
  client_id     = var.keycloak_client_id
  client_secret = var.keycloak_client_secret
}
