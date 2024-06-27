variable "snowflake_password" {
  type = string
  sensitive = true
}

resource "propel_data_source" "my_data_source" {
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

variable "http_basic_auth_password" {
  type = string
  sensitive = true
}

resource "propel_data_source" "my_webhook_data_source" {
  unique_name = "My Webhook Data Source"
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

variable "kafka_password" {
  type = string
  sensitive = true
}

resource "propel_data_source" "my_kafka_data_source" {
  unique_name = "My Kafka Data Source"
  description = "This is an example of a Kafka Data Source"
  type        = "KAFKA"
  kafka_connection_settings {
    auth = "SCRAM-SHA-256"
    user = "user"
    password = var.kafka_password
    tls = true
    bootstrap_servers = ["localhost:9092"]
  }
}

variable "clickhouse_password" {
  type = string
  sensitive = true
}

resource "propel_data_source" "my_clickhouse_data_source" {
  unique_name = "My ClickHouse Data Source"
  description = "This is an example of a ClickHouse Data Source"
  type        = "CLICKHOUSE"
  clickhouse_connection_settings {
    url = "http://127.0.0.1:8123"
    user = "user"
    password = var.clickhouse_password
    database = "sample"
  }
}
