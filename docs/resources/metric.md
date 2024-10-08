---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "propel_metric Resource - propel"
subcategory: ""
description: |-
  Provides a Propel Metric resource. This can be used to create and manage Propel Metrics.
---

# propel_metric (Resource)

Provides a Propel Metric resource. This can be used to create and manage Propel Metrics.

## Example Usage

```terraform
resource "propel_metric" "my_sum_metric" {
  unique_name = "my_sum_metric"
  description = "This is an example of a Sum Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type    = "SUM"
  measure = "price"

  filter {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  filter {
    column   = "country"
    operator = "EQUALS"
    value    = "bar"
    or = jsonencode([{"column":"country","operator":"EQUALS","value":"baz"}])
  }

  dimensions = ["store"]
}

resource "propel_metric" "my_count_metric" {
  unique_name = "my_count_metric"
  description = "This is an example of a Count Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type = "COUNT"

  filter {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  dimensions = ["store"]
}

resource "propel_metric" "my_count_distinct_metric" {
  unique_name = "my_count_distinct_metric"
  description = "This is an example of a Count Distinct Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type      = "COUNT_DISTINCT"
  dimension = "product_id"

  filter {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  dimensions = ["store"]
}


resource "propel_metric" "my_average_metric" {
  unique_name = "my_average_metric"
  description = "This is an example of a Average Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type      = "AVERAGE"
  measure   = "price"

  filter {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  dimensions = ["store"]
}

resource "propel_metric" "my_min_metric" {
  unique_name = "my_min_metric"
  description = "This is an example of a Min Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type      = "MIN"
  measure   = "price"

  filter {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  dimensions = ["store"]
}

resource "propel_metric" "my_max_metric" {
  unique_name = "my_max_metric"
  description = "This is an example of a Max Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type      = "MAX"
  measure   = "price"

  filter {
    column   = "product_name"
    operator = "EQUALS"
    value    = "foo"
  }

  dimensions = ["store"]
}

resource "propel_metric" "my_custom_metric" {
  unique_name = "my_custom_metric"
  description = "This is an example of a Custom Metric"
  data_pool   = propel_data_pool.my_data_pool.id

  type         = "CUSTOM"
  expression   = "SUM(price * quantity) / COUNT()"

  filter {
    column   = "price"
    operator = "IS_NOT_NULL"
  }

  dimensions = ["store"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data_pool` (String) The Data Pool that powers this Metric.
- `type` (String) The Metric type. The different Metric types determine how the values are calculated.

### Optional

- `access_control_enabled` (Boolean, Deprecated) Whether or not access control is enabled for the Metric.
- `description` (String) The Metric's description.
- `dimension` (String) The Dimension where the count distinct operation is going to be performed. Only valid for COUNT_DISTINCT Metrics.
- `dimensions` (List of String) The Metric's Dimensions. These Dimensions are available to Query Filters.
- `expression` (String) The custom expression for aggregating data in a Metric. Only valid for CUSTOM Metrics.
- `filter` (Block List) Metric Filters allow defining a Metric with a subset of records from the given Data Pool. If no Metric Filters are present, all records will be included. To filter at query time, add Dimensions and use the `filters` property on the `timeSeriesInput`, `counterInput`, or `leaderboardInput` objects. There is no need to add `filters` to be able to filter at query time. (see [below for nested schema](#nestedblock--filter))
- `measure` (String) The Dimension to be summed, taken the minimum of, taken the maximum of, averaged, etc. Only valid for SUM, MIN, MAX and AVERAGE Metrics.
- `unique_name` (String) The Metric's name.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

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
terraform import propel_metric.my_metric MET00000000000000000000000000
```
