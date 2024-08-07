resource "propel_materialized_view" "my_materialized_view" {
  unique_name = "My materialized view"
  description = "This is an example of a Materialized View"
  sql = "SELECT date, account_id FROM ${propel_data_pool.my_data_pool.id}"

  new_data_pool {
    unique_name = "My MV destination Data Pool"
    timestamp = "date"
    access_control_enabled = true

    table_settings {
      engine {
        type = "MERGE_TREE"
      }
      partition_by = ["toYYYYMM(date)"]
      order_by = ["date", "account_id"]
    }
  }

  backfill = true
}

resource "propel_materialized_view" "existing_data_pool_mv" {
  unique_name = "My existing Data Pool materialized view"
  description = "This is an example of a Materialized View"
  sql = "SELECT date, account_id FROM ${propel_data_pool.my_data_pool.id}"

  existing_data_pool {
    id = "${propel_data_pool.my_existing_data_pool.id}"
  }
}