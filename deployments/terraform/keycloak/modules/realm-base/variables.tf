variable "realm_name" {
  type = string
}

variable "enabled" {
  type    = bool
  default = true
}

variable "display_name" {
  type = string
}

variable "registration_allowed" {
  type    = bool
  default = false
}

variable "login_with_email_allowed" {
  type    = bool
  default = true
}

variable "duplicate_emails_allowed" {
  type    = bool
  default = false
}

variable "reset_password_allowed" {
  type    = bool
  default = true
}

variable "edit_username_allowed" {
  type    = bool
  default = false
}

variable "ssl_required" {
  type    = string
  default = "external"
}
