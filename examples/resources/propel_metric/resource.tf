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
