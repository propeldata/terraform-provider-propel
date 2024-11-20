terraform {
  required_providers {
    propel = {
      source  = "propeldata/propel"
      version = "1.3.5"
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
  type        = "HTTP"

  http_connection_settings {
    basic_auth {
      username = "foo"
      password = var.http_basic_auth_password
    }
  }
}

resource "propel_data_source" "my_webhook_data_source" {
  unique_name = "MyWebhookDataPool"
  description = "This is an example of a Webhook Data Source"
  type        = "WEBHOOK"

  webhook_connection_settings {
    timestamp = "event_timestamp"
    unique_id = "event_id"
    tenant = "customer_id"

    column {
      name = "event_id"
      type = "STRING"
      nullable = false
      json_property = "event_id"
    }

    column {
      name = "customer_id"
      type = "STRING"
      nullable = false
      json_property = "customer_id"
    }

    column {
      name = "event_timestamp"
      type = "TIMESTAMP"
      nullable = false
      json_property = "event_timestamp"
    }

    column {
      name = "customer_name"
      type = "STRING"
      nullable = true
      json_property = "customer_name"
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

resource "propel_materialized_view" "materialized_view" {
  unique_name = "My Materialized View"
  sql = "SELECT customer_id, value, timestamp FROM \"${propel_data_pool.data_pool.id}\""

  new_data_pool {
    unique_name = "My SummingMergeTree Data Pool"
    timestamp = "timestamp"
    unique_id = "customer_id"
    access_control_enabled = true
    table_settings {
      engine {
        type = "SUMMING_MERGE_TREE"
        columns = ["value"]
      }
    }
  }
  backfill = true
}

resource "propel_materialized_view" "materialized_view_existing_data_pool" {
  unique_name = "My Materialized View"
  sql = "SELECT customer_id, value, timestamp FROM \"Sales\""
  existing_data_pool {
    id = "${propel_data_pool.data_pool.id}"
  }
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

output "materialized_view_id" {
  value = propel_materialized_view.materialized_view.id
}
