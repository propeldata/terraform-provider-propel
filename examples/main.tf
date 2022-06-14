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

resource "propel_datasource" "datasource" {
  unique_name = var.datasource_unique_name
  description = var.datasource_description

  connection_settings {
    username = var.datasource_username
    password = var.datasource_password
    warehouse = var.datasource_warehouse
    role = var.datasource_role
    account = var.datasource_account
    database = var.datasource_database
    schema = var.datasource_schema
  }
}

resource "propel_datapool" "datapool" {
  unique_name = var.datapool_unique_name
  description = var.datapool_description
  datasource = propel_datasource.datasource.id
  table = var.datapool_table
  timestamp = var.datapool_timestamp
}

resource "propel_metric" "count_metric" {
  unique_name = var.metric_unique_name
  description = var.metric_description
  datapool = propel_datapool.datapool.id
  type = "COUNT"

  filter {
    column = var.metric_filter_column
    operator = var.metric_filter_operator
    value = var.metric_filter_value
  }

  dimensions = var.metric_dimensions
}

resource "propel_metric" "sum_metric" {
  unique_name = var.metric_sum_unique_name
  datapool = propel_datapool.datapool.id
  type = "SUM"
  measure = var.metric_sum_measure

  filter {
    column = var.metric_filter_column
    operator = var.metric_filter_operator
    value = var.metric_filter_value
  }

  dimensions = var.metric_dimensions
}

resource "propel_metric" "count_distinct_metric" {
  unique_name = var.metric_count_distinct_unique_name
  datapool = propel_datapool.datapool.id
  type = "COUNT_DISTINCT"
  dimension = var.metric_count_distinct_dimension

  filter {
    column = var.metric_filter_column
    operator = var.metric_filter_operator
    value = var.metric_filter_value
  }

  dimensions = var.metric_dimensions
}

output "datasource_id" {
  value = propel_datasource.datasource.id
}

output "datapool_id" {
  value = propel_datapool.datapool.id
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
