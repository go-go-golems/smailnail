variable "realm_id" {
  type = string
}

variable "client_id" {
  type = string
}

variable "name" {
  type = string
}

variable "enabled" {
  type    = bool
  default = true
}

variable "access_type" {
  type    = string
  default = "PUBLIC"
}

variable "client_secret" {
  type      = string
  default   = null
  sensitive = true
  nullable  = true
}

variable "direct_access_grants_enabled" {
  type    = bool
  default = false
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
  default = []
}

variable "optional_scopes" {
  type    = list(string)
  default = []
}

variable "manage_scope_attachments" {
  type    = bool
  default = true
}
