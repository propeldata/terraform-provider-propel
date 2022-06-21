---
page_title: "propel_metric Resource - terraform-provider-propel"
subcategory: ""
description: |-
Provides a Propel metric resource. This can be used to create and manage Propel metrics.
---

# Resource `propel_metric`
Provides a Propel Metric resource. This can be used to create and manage Propel Metrics.

## Example Usage

```terraform
resource "propel_metric" "my_sum_metric" {
  unique_name = "my_sum_metric"
  description = "Metric Description"
  data_pool = propel_data_pool.my_data_pool.id
  
  type = "SUM"
  measure = "price"
  
  filter {
    column = "product_name"
    operator = "EQUALS"
    value = "foo"
  }

  filter {
    column = "country"
    operator = "EQUALS"
    value = "bar"
  }
  
  dimensions = ["store"]
}

resource "propel_metric" "my_count_metric" {
  unique_name = "my_count_metric"
  description = "Metric Description"
  data_pool = propel_data_pool.my_data_pool.id
  
  type = "COUNT"

  filter {
    column = "product_name"
    operator = "EQUALS"
    value = "foo"
  }

  dimensions = ["store"]
}

resource "propel_metric" "my_count_distinct_metric" {
  unique_name = "my_count_distinct_metric"
  description = "Metric Description"
  data_pool = propel_data_pool.my_data_pool.id
  
  type = "COUNT_DISTINCT"
  dimension = "product_id"
  
  filter {
    column = "product_name"
    operator = "EQUALS"
    value = "foo"
  }

  dimensions = ["store"]
}
```

## Schema

### Required
- `type` - (String) The type of Metric you want to create

### Optional
- `unique_name` - (String) The unique name of the Metric.
- `description` - (String) The description of the Metric.
- `dimension` - (String) The column on which the count distinct is going to be performed.
- `measure` - (String) The column you want to sum.
- `dimensions` - (List of String) An array of column names that are used as dimensions for your Metric.

### Nested schema for `filter`
Filters allow defining a Metric with a subset of records from the given Data Pool. If no filters are present, all records will be included.

### Required
- `column` (String) The column name.
- `database` (String) The operator for the filter. can be `EQUALS`, `NOT_EQUALS`, `GREATER_THAN`, `GREATER_THAN_OR_EQUAL_TO`,  `LESS_THAN` and `LESS_THAN_OR_EQUAL_TO`.
- `value` (String) The column value.

## Import
Import is supported using the following syntax:
```
terraform import propel_metric.my_sum_meric MET00000000000000000000000000
```
