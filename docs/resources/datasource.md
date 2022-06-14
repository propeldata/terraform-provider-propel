---
page_title: "propel_datasource Resource - terraform-provider-propel"
subcategory: ""
description: |-
  The datasource resource allows you to configure a Propel DataSource.
---

# Resource `propel_datasource`
Provides a Propel DataSource resource. This can be used to create and manage Propel DataSources.

## Example Usage

```terraform
resource "propel_datasource" "my_datasource" {
  unique_name = "my_datasource"
  description = "description"

  connection_settings {
    account = "snowflake-account"
    database = "snowflake-database"
    warehouse = "snowflake-warehouse"
    schema = "snowflake-schema"
    role = "snowflake-role"
    username = "snowflake-username"
    password = "snowflake-password"
  }
}
```

## Schema

### Optional
- `unique_name` - (String) Give your Data Source a unique name
- `description` - (String) Give your Data Source a description

### Nested schema for `connection_settings`
Snowflake connection details

### Required
- `account` (String) The Snowflake account identifier 
- `database` (String) The Snowflake database
- `warehouse` (String) The Snowflake warehouse
- `schema` (String) The Snowflake schema
- `role` (String) The Snowflake role
- `username` (String) The Snowflake username
- `password` (String, Sensitive) The Snowflake password

## Import
Import is supported using the following syntax:
`terraform import propel_datasource.my_datasource DSO1111111111111111111111111111`


