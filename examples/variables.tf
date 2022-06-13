variable "client_id" { type = string }
variable "client_secret" {
  type = string
  sensitive = true
}

variable "datasource_unique_name" { type = string }
variable "datasource_description" { type = string }
variable "datasource_username" { type = string }
variable "datasource_password" {
  type = string
  sensitive = true
}
variable "datasource_warehouse" { type = string }
variable "datasource_role" { type = string }
variable "datasource_account" { type = string }
variable "datasource_database" { type = string }
variable "datasource_schema" { type = string }

variable "datapool_unique_name" { type = string }
variable "datapool_description" { type = string }
variable "datapool_table" { type = string }
variable "datapool_timestamp" { type = string }

variable "metric_unique_name" { type = string }
variable "metric_description" { type = string }
variable "metric_filter_column" { type = string }
variable "metric_filter_operator" { type = string }
variable "metric_filter_value" { type = string }
variable "metric_dimensions" { type = set(string) }
variable "metric_sum_unique_name" { type = string }
variable "metric_sum_measure" { type = string }
variable "metric_count_distinct_unique_name" { type = string }
variable "metric_count_distinct_dimension" { type = string }
