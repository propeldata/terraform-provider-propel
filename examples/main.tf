terraform {
  required_providers {
    propel = {
      source  = "propeldata/propel"
      version = "0.0.3"
    }
  }
}

provider "propel" {
  client_id     = var.client_id
  client_secret = var.client_secret
}

resource "propel_data_source" "snowflake_data_source" {
  unique_name = "My Snowflake Data Source"
  description = "This is an example of a Snowflake Data Source"
  type        = "SNOWFLAKE"

  snowflake_connection_settings {
    account   = "Snowflake Account"
    database  = "Snowflake Database"
    warehouse = "Snowflake Warehouse"
    schema    = "Snowflake Schema"
    role      = "Snowflake Role"
    username  = "Snowflake Username"
    password  = var.snowflake_password
  }
}

resource "propel_data_source" "http_data_source" {
  unique_name = "My HTTP Data Source"
  description = "This is an example of an HTTP Data Source"
  type        = "Http"
}

resource "propel_data_pool" "data_pool" {
  unique_name = "My Data Pool"
  description = "Data Pool Description"
  data_source = propel_data_source.snowflake_data_source.id

  table     = "sample"
  timestamp = "date"

  column {
    name     = "date"
    type     = "TIMESTAMP"
    nullable = false
  }
}

resource "propel_metric" "count_metric" {
  unique_name = "My Count Metric"
  description = "Metric Description"
  data_pool   = propel_data_pool.data_pool.id

  type = "COUNT"

  filter {
    column   = "column_3"
    operator = "EQUALS"
    value    = "value"
  }

  filter {
    column   = "column_4"
    operator = "EQUALS"
    value    = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "sum_metric" {
  unique_name = "My Sum Metric"
  data_pool   = propel_data_pool.data_pool.id

  type    = "SUM"
  measure = "column_1"

  filter {
    column   = "column_3"
    operator = "EQUALS"
    value    = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "count_distinct_metric" {
  unique_name = "My Count Distinct Metric"
  data_pool   = propel_data_pool.data_pool.id

  type      = "COUNT_DISTINCT"
  dimension = "column_1"

  filter {
    column   = "column_4"
    operator = "EQUALS"
    value    = "value"
  }

  dimensions = ["column_1", "column_2"]
}

output "snowflake_data_source_id" {
  value = propel_data_source.snowflake_data_source.id
}

output "http_data_source_id" {
  value = propel_data_source.http_data_source.id
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
