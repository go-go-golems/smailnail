variable "realm_id" {
  type = string
}

variable "create_alice" {
  type    = bool
  default = true
}

variable "alice_username" {
  type    = string
  default = "alice"
}

variable "alice_email" {
  type    = string
  default = "alice@example.com"
}

variable "alice_first_name" {
  type    = string
  default = "Alice"
}

variable "alice_last_name" {
  type    = string
  default = "Example"
}

variable "alice_password" {
  type      = string
  default   = "secret"
  sensitive = true
}
