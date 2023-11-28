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