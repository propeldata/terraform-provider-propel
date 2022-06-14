---
page_title: "propel_datapool Resource - terraform-provider-propel"
subcategory: ""
description: |-
The datapool resource allows you to configure a Propel DataPool.
---

# Resource `propel_datapool`
Provides a Propel DataPool resource. This can be used to create and manage Propel DataPools.

## Example Usage

```terraform
resource "propel_datapool" "my_datapool" {
  unique_name = "my_datapool"
  description = "description"
  table = "events"
  timestamp = "date"
  datasource = propel_datasource.my_datasource.id
}
```

## Schema

### Required
- `table` - (String) The name of the table you want to use
- `timestamp` - (String) The primary timestamp column to use

### Optional
- `unique_name` - (String) Give your Data Pool a unique name
- `description` - (String) Give your Data Pool a description

## Import
Import is supported using the following syntax:
`terraform import propel_datapool.my_datapool DPO111111111111111111111111111`


