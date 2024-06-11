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
  unique_name = "My Webhook Data Pool"
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
