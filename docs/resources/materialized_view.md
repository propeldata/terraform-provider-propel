---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "propel_materialized_view Resource - propel"
subcategory: ""
description: |-
  Provides a Propel Materialized View resource. This can be used to create and manage Propel Materialized Views.
---

# propel_materialized_view (Resource)

Provides a Propel Materialized View resource. This can be used to create and manage Propel Materialized Views.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `sql` (String) The SQL that the Materialized View executes.

### Optional

- `backfill` (Boolean) Whether historical data should be backfilled or not.
- `description` (String) The Materialized View's description.
- `existing_data_pool` (Block List, Max: 1) If specified, the Materialized View will target an existing Data Pool. Ensure the Data Pool's schema is compatible with your Materialized View's SQL statement. (see [below for nested schema](#nestedblock--existing_data_pool))
- `new_data_pool` (Block List, Max: 1) If specified, the Materialized View will create and target a new Data Pool. You can further customize the new Data Pool's engine settings. (see [below for nested schema](#nestedblock--new_data_pool))
- `unique_name` (String) The Materialized View's name.

### Read-Only

- `account` (String) The Materialized View's Account.
- `destination` (String) The Materialized View's destination (AKA "target") Data Pool.
- `environment` (String) The Environment that the Materialized View belongs to.
- `id` (String) The ID of this resource.
- `others` (Set of String) Other Data Pools queried by the Materialized View.
- `source` (String) The Materialized View's source Data Pool.

<a id="nestedblock--existing_data_pool"></a>
### Nested Schema for `existing_data_pool`

Required:

- `id` (String) The ID of the Data Pool.


<a id="nestedblock--new_data_pool"></a>
### Nested Schema for `new_data_pool`

Optional:

- `access_control_enabled` (Boolean) Enables or disables access control for the Data Pool. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.
- `description` (String) The Data Pool's description.
- `table_settings` (Block List, Max: 1) Override the Data Pool's table settings. These describe how the Data Pool's table is created in ClickHouse, and a default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if any. You can override these defaults in order to specify a custom table engine, custom ORDER BY, etc. (see [below for nested schema](#nestedblock--new_data_pool--table_settings))
- `timestamp` (String) Optionally specify the Data Pool's primary timestamp. This will influence the Data Pool's engine settings.
- `unique_name` (String) The Data Pool's unique name.

<a id="nestedblock--new_data_pool--table_settings"></a>
### Nested Schema for `new_data_pool.table_settings`

Optional:

- `engine` (Block List, Max: 1) The ClickHouse table engine for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified. (see [below for nested schema](#nestedblock--new_data_pool--table_settings--engine))
- `order_by` (Set of String) The ORDER BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.
- `partition_by` (Set of String) The PARTITION BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.
- `primary_key` (Set of String) The PRIMARY KEY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.

<a id="nestedblock--new_data_pool--table_settings--engine"></a>
### Nested Schema for `new_data_pool.table_settings.engine`

Optional:

- `columns` (Set of String) The columns argument for the SummingMergeTree table engine.
- `type` (String)
- `ver` (String) The `ver` parameter to the ReplacingMergeTree table engine.

## Import

Import is supported using the following syntax:

```shell
terraform import propel_materialized_view.my_materialized_view MAT00000000000000000000000000
```