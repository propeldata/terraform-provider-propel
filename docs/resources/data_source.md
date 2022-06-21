---
page_title: "propel_data_source Resource - terraform-provider-propel"
subcategory: ""
description: |-
  The data source resource allows you to configure a Propel Data Source.
---

# Resource `propel_data_source`
Provides a Propel Data Source resource. This can be used to create and manage Propel Data Sources.

## Example Usage

```terraform
variable "snowflake_password" {
  type = string
  sensitive = true
}

resource "propel_data_source" "my_data_source" {
  unique_name = "My Data Source"
  description = "Data Source Description"

  connection_settings {
    account = "snowflake-account"
    database = "snowflake-database"
    warehouse = "snowflake-warehouse"
    schema = "snowflake-schema"
    role = "snowflake-role"
    username = "snowflake-username"
    password = var.snowflake_password
  }
}
```

## Schema

### Optional
- `unique_name` - (String) Give your Data Source a unique name.
- `description` - (String) Give your Data Source a description.

### Nested schema for `connection_settings`
Snowflake connection details.

### Required
- `account` (String) The Snowflake account identifier.
- `database` (String) The Snowflake database.
- `warehouse` (String) The Snowflake warehouse.
- `schema` (String) The Snowflake schema.
- `role` (String) The Snowflake role.
- `username` (String) The Snowflake username.
- `password` (String, Sensitive) The Snowflake password.

## Import
Import is supported using the following syntax:
```
terraform import propel_data_source.my_data_source DSO00000000000000000000000000
```
