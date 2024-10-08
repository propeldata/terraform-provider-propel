---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "propel_data_pool_access_policy Resource - propel"
subcategory: ""
description: |-
  Provides a Propel Data Pool Access Policy resource. This can be used to create and manage Propel Data Pool Access Policies.
---

# propel_data_pool_access_policy (Resource)

Provides a Propel Data Pool Access Policy resource. This can be used to create and manage Propel Data Pool Access Policies.

## Example Usage

```terraform
resource "propel_data_pool_access_policy" "my_data_pool_access_policy" {
  unique_name = "My Data Pool Access Policy"
  description = "This is an example of a Data Pool Access Policy"
  data_pool = propel_data_source.my_data_pool.id

  columns = ["*"]

  row {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  row {
    column   = "country"
    operator = "EQUALS"
    value    = "bar"
    or = jsonencode([{"column":"country","operator":"EQUALS","value":"baz"}])
  }

  applications = ["APP00000000000000000000000000"]

}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `columns` (List of String) The list of columns that the Access Policy makes available for querying. Set "*" to allow all columns.
- `data_pool` (String) The Data Pool to which this Access Policy belongs.

### Optional

- `applications` (Set of String) The list of applications to which the Access Policy is assigned.
- `description` (String) The Data Pool Access Policy's description.
- `row` (Block List) Row-level filters that the Access Policy applies before executing queries. Not setting any row filters means all rows can be queried. (see [below for nested schema](#nestedblock--row))
- `unique_name` (String) The Data Pool Access Policy's name.

### Read-Only

- `account` (String) The Account to which the Data Pool Access Policy belongs.
- `environment` (String) The Environment to which the Data Pool Access Policy belongs.
- `id` (String) The ID of this resource.

<a id="nestedblock--row"></a>
### Nested Schema for `row`

Required:

- `column` (String) The name of the column to filter on.
- `operator` (String) The operation to perform when comparing the column and filter values.

Optional:

- `and` (String) Additional filters to AND with this one. AND takes precedence over OR. It is defined as a JSON string value.
- `or` (String) Additional filters to OR with this one. AND takes precedence over OR. It is defined as a JSON string value.
- `value` (String) The value to compare the column to.

## Import

Import is supported using the following syntax:

```shell
terraform import propel_data_pool_access_policy.my_data_pool_access_policy POL00000000000000000000000000
```
