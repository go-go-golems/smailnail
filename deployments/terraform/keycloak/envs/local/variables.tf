variable "keycloak_url" {
  type = string
}

variable "keycloak_base_path" {
  type    = string
  default = ""
}

variable "keycloak_admin_realm" {
  type    = string
  default = "master"
}

variable "keycloak_client_id" {
  type    = string
  default = "admin-cli"
}

variable "keycloak_client_secret" {
  type      = string
  default   = null
  sensitive = true
  nullable  = true
}

variable "keycloak_username" {
  type    = string
  default = "admin"
}

variable "keycloak_password" {
  type      = string
  default   = "admin"
  sensitive = true
}

variable "web_client_secret" {
  type      = string
  sensitive = true
}

variable "alice_password" {
  type      = string
  default   = "secret"
  sensitive = true
}

variable "realm_name" {
  type    = string
  default = "smailnail-dev"
}

variable "realm_display_name" {
  type    = string
  default = "smailnail-dev"
}
