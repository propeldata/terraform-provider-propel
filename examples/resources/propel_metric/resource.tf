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