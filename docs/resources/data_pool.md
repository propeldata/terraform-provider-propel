---
page_title: "propel_data_pool Resource - terraform-provider-propel"
subcategory: ""
description: |-
The data pool resource allows you to configure a Propel Data Pool.
---

# Resource `propel_data_pool`
Provides a Propel Data Pool resource. This can be used to create and manage Propel Data Pools.

## Example Usage

```terraform
resource "propel_data_pool" "my_data_pool" {
  unique_name = "My Data Pool"
  description = "Data Pool Description"
  table = "events"
  timestamp = "date"
  data_source = propel_data_source.my_data_source.id
}
```

## Schema

### Required
- `table` - (String) The name of the table you want to use.
- `timestamp` - (String) The primary timestamp column to use.

### Optional
- `unique_name` - (String) Give your Data Pool a unique name.
- `description` - (String) Give your Data Pool a description.

## Import
Import is supported using the following syntax:
```
terraform import propel_data_pool.my_data_pool DPO00000000000000000000000000
```
