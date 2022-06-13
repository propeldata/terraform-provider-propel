variable "client_id" { type = string }
variable "client_secret" {
  type = string
  sensitive = true
}

variable "datasource_unique_name" { type = string }
variable "datasource_description" { type = string }
variable "datasource_username" { type = string }
variable "datasource_password" { type = string }
variable "datasource_warehouse" { type = string }
variable "datasource_role" { type = string }
variable "datasource_account" { type = string }
variable "datasource_database" { type = string }
variable "datasource_schema" { type = string }

variable "datapool_unique_name" { type = string }
variable "datapool_description" { type = string }
variable "datapool_table" { type = string }
variable "datapool_timestamp" { type = string }
