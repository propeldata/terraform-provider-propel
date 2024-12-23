---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "propel_data_pool Resource - propel"
subcategory: ""
description: |-
  Provides a Propel Data Pool resource. This can be used to create and manage Propel Data Pools.
---

# propel_data_pool (Resource)

Provides a Propel Data Pool resource. This can be used to create and manage Propel Data Pools.

## Example Usage

```terraform
resource "propel_data_pool" "my_data_pool" {
  unique_name = "My Data Pool"
  description = "This is an example of a Data Pool"
  data_source = propel_data_source.my_data_source.id
  table       = "events"
  timestamp   = "date"

  access_control_enabled = false

  column {
    name     = "date"
    type     = "TIMESTAMP"
    nullable = false
  }
  column {
    name     = "account_id"
    type     = "STRING"
    nullable = false
  }
  tenant_id = "account_id"

  syncing {
    interval = "EVERY_1_HOUR"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `access_control_enabled` (Boolean) Whether the Data Pool has access control enabled or not. If the Data Pool has access control enabled, Applications must be assigned Data Pool Access Policies in order to query the Data Pool and its Metrics.
- `column` (Block List) The list of columns, their types and nullability. (see [below for nested schema](#nestedblock--column))
- `data_source` (String) The Data Source that the Data Pool belongs to.
- `description` (String) The Data Pool's description.
- `syncing` (Block List, Max: 1) The Data Pool's syncing settings. (see [below for nested schema](#nestedblock--syncing))
- `table` (String) The name of the Data Pool's table.
- `table_settings` (Block List, Max: 1) Override the Data Pool's table settings. These describe how the Data Pool's table is created in ClickHouse, and a default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if any. You can override these defaults in order to specify a custom table engine, custom ORDER BY, etc. (see [below for nested schema](#nestedblock--table_settings))
- `tenant_id` (String, Deprecated) The tenant ID for restricting access between customers.
- `timestamp` (String) The Data Pool's timestamp column.
- `unique_id` (String, Deprecated) The Data Pool's unique ID column. Propel uses the primary timestamp and a unique ID to compose a primary key for determining whether records should be inserted, deleted, or updated within the Data Pool. Only for Snowflake Data Pools.
- `unique_name` (String) The Data Pool's name.

### Read-Only

- `account` (String) The Account that the Data Pool belongs to.
- `environment` (String) The Environment that the Data Pool belongs to.
- `id` (String) The ID of this resource.
- `status` (String) The Data Pool's status.

<a id="nestedblock--column"></a>
### Nested Schema for `column`

Required:

- `name` (String) The column name.
- `nullable` (Boolean) Whether the column's type is nullable or not.
- `type` (String) The column type.

Optional:

- `clickhouse_type` (String) The ClickHouse type to use when `type` is set to `CLICKHOUSE`.


<a id="nestedblock--syncing"></a>
### Nested Schema for `syncing`

Required:

- `interval` (String) The syncing interval.

Read-Only:

- `last_synced_at` (String) The date and time of the most recent Sync in UTC.
- `status` (String) Indicates whether syncing is enabled or disabled.


<a id="nestedblock--table_settings"></a>
### Nested Schema for `table_settings`

Optional:

- `engine` (Block List, Max: 1) The ClickHouse table engine for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified. (see [below for nested schema](#nestedblock--table_settings--engine))
- `order_by` (List of String) The ORDER BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.
- `partition_by` (List of String) The PARTITION BY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.
- `primary_key` (List of String) The PRIMARY KEY clause for the Data Pool's table. This field is optional. A default will be chosen based on the Data Pool's `timestamp` and `uniqueId` values, if specified.

<a id="nestedblock--table_settings--engine"></a>
### Nested Schema for `table_settings.engine`

Optional:

- `columns` (List of String) The columns argument for the SummingMergeTree table engine.
- `type` (String) The ClickHouse table engine.
- `ver` (String) The `ver` parameter to the ReplacingMergeTree table engine.

## Import

Import is supported using the following syntax:

```shell
terraform import propel_data_pool.my_data_pool DPO00000000000000000000000000
```
