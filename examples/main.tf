terraform {
  required_providers {
    propel = {
      version = ">= 0.0.1"
      source = "propeldata.com/propeldata/propel"
    }
  }
}

provider "propel" {
  client_id = var.client_id
  client_secret = var.client_secret
}

resource "propel_data_source" "data_source" {
  unique_name = "My Data Source"
  description = "Data Source Description"

  connection_settings {
    account = "Snowflake Account"
    database = "Snowflake Database"
    warehouse = "Snowflake Warehouse"
    schema = "Snowflake Schema"
    role = "Snowflake Role"
    username = "Snowflake Username"
    password = var.snowflake_password
  }
}

resource "propel_data_pool" "data_pool" {
  unique_name = "My Data Pool"
  description = "Data Pool Description"
  datasource = propel_data_source.data_source.id

  table = "sample"
  timestamp = "date"
}

resource "propel_metric" "count_metric" {
  unique_name = "My Count Metric"
  description = "Metric Description"
  datapool = propel_data_pool.data_pool.id

  type = "COUNT"

  filter {
    column = "column_3"
    operator = "EQUALS"
    value = "value"
  }

  filter {
    column = "column_4"
    operator = "EQUALS"
    value = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "sum_metric" {
  unique_name = "My Sum Metric"
  datapool = propel_data_pool.data_pool.id

  type = "SUM"
  measure = "column_1"

  filter {
    column = "column_3"
    operator = "EQUALS"
    value = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "count_distinct_metric" {
  unique_name = "My Count Distinct Metric"
  datapool = propel_data_pool.data_pool.id

  type = "COUNT_DISTINCT"
  dimension = "column_1"

  filter {
    column = "column_4"
    operator = "EQUALS"
    value = "value"
  }

  dimensions = ["column_1", "column_2"]
}

output "data_source_id" {
  value = propel_data_source.data_source.id
}

output "data_pool_id" {
  value = propel_data_pool.data_pool.id
}

output "count_metric_id" {
  value = propel_metric.count_metric.id
}

output "sum_metric_id" {
  value = propel_metric.sum_metric.id
}

output "count_distinct_metric_id" {
  value = propel_metric.count_distinct_metric.id
}
