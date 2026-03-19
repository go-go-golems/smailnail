terraform {
  required_providers {
    keycloak = {
      source = "keycloak/keycloak"
    }
  }
}

resource "keycloak_user" "alice" {
  count          = var.create_alice ? 1 : 0
  realm_id       = var.realm_id
  username       = var.alice_username
  enabled        = true
  email          = var.alice_email
  email_verified = true
  first_name     = var.alice_first_name
  last_name      = var.alice_last_name

  initial_password {
    value     = var.alice_password
    temporary = false
  }
}
