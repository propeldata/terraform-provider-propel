terraform {
  required_providers {
    propel = {
      source  = "propeldata/propel"
      version = "1.2.0"
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

  http_connection_settings {
    basic_auth {
      username = "foo"
      password = var.http_basic_auth_password
    }
  }
}

resource "propel_data_source" "webhook_data_source" {
  unique_name = "My Webhook Data Source"
  description = "This is an example of a Webhook Data Source"
  type        = "Webhook"

  webhook_connection_settings {
    timestamp = "date"
    unique_id = "id"

    column {
      name = "id"
      type = "STRING"
      nullable = false
      json_property = "id"
    }

    column {
      name = "customer_id"
      type = "STRING"
      nullable = false
      json_property = "customer_id"
    }

    basic_auth {
      username = "foo"
      password = var.http_basic_auth_password
    }
  }
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

resource "propel_data_pool_access_policy" "data_pool_access_policy" {
  unique_name = "My Data Pool Access Policy"
  description = "Data Pool Access Policy Description"
  data_pool = propel_data_pool.data_pool.id

  columns = ["*"]

  row {
    column   = "column_0"
    operator = "EQUALS"
    value    = "value"
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

resource "propel_metric" "average_metric" {
  unique_name = "My Average Metric"
  data_pool   = propel_data_pool.data_pool.id

  type      = "AVERAGE"
  measure   = "column_1"

  filter {
    column   = "column_3"
    operator = "EQUALS"
    value    = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "min_metric" {
  unique_name = "My MIN Metric"
  data_pool   = propel_data_pool.data_pool.id

  type      = "MIN"
  measure   = "column_1"

  filter {
     column   = "column_3"
     operator = "EQUALS"
     value    = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "max_metric" {
  unique_name = "My MAX Metric"
  data_pool   = propel_data_pool.data_pool.id

  type      = "MAX"
  measure   = "column_1"

  filter {
    column   = "column_3"
    operator = "EQUALS"
    value    = "value"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_metric" "custom_metric" {
  unique_name = "My CUSTOM Metric"
  data_pool   = propel_data_pool.data_pool.id

  type         = "CUSTOM"
  expression   = "SUM(column_1 * column_2) / COUNT()"

  filter {
    column   = "column_1"
    operator = "IS_NOT_NULL"
  }

  dimensions = ["column_1", "column_2"]
}

resource "propel_policy" "sum_metric_policy" {
  type = "ALL_ACCESS"
  metric = propel_metric.sum_metric.id
  application = var.client_id
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

output "average_metric_id" {
  value = propel_metric.average_metric.id
}

output "min_metric_id" {
  value = propel_metric.min_metric.id
}

output "max_metric_id" {
  value = propel_metric.max_metric.id
}

output "custom_metric_id" {
  value = propel_metric.custom_metric.id
}

output "sum_metric_policy_id" {
  value = propel_policy.sum_metric_policy.id
}