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
  type = string
}

variable "keycloak_client_secret" {
  type      = string
  sensitive = true
}

variable "web_client_secret" {
  type      = string
  sensitive = true
}

variable "mcp_client_secret" {
  type      = string
  default   = null
  sensitive = true
  nullable  = true
}
