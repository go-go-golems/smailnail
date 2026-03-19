terraform {
  required_providers {
    keycloak = {
      source = "keycloak/keycloak"
    }
  }
}

resource "keycloak_realm" "this" {
  realm                       = var.realm_name
  enabled                     = var.enabled
  display_name                = var.display_name
  registration_allowed        = var.registration_allowed
  login_with_email_allowed    = var.login_with_email_allowed
  duplicate_emails_allowed    = var.duplicate_emails_allowed
  reset_password_allowed      = var.reset_password_allowed
  edit_username_allowed       = var.edit_username_allowed
  ssl_required                = var.ssl_required
  default_signature_algorithm = var.default_signature_algorithm
}
