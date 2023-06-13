variable "client_id" {
  type = string
}

variable "client_secret" {
  type      = string
  sensitive = true
}

variable "snowflake_password" {
  type      = string
  sensitive = true
}

variable "http_basic_auth_password" {
  type      = string
  sensitive = true
}