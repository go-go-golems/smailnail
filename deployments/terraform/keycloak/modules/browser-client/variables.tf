variable "realm_id" {
  type = string
}

variable "client_id" {
  type = string
}

variable "name" {
  type     = string
  default  = null
  nullable = true
}

variable "client_secret" {
  type      = string
  sensitive = true
}

variable "enabled" {
  type    = bool
  default = true
}

variable "use_refresh_tokens" {
  type    = bool
  default = true
}

variable "valid_redirect_uris" {
  type = list(string)
}

variable "web_origins" {
  type = list(string)
}

variable "default_scopes" {
  type    = list(string)
  default = ["profile", "email"]
}

variable "optional_scopes" {
  type    = list(string)
  default = []
}

variable "manage_scope_attachments" {
  type    = bool
  default = true
}
