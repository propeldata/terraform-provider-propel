resource "propel_data_pool" "my_data_pool" {
  unique_name = "My Data Pool"
  description = "This is an example of a Data Pool"
  data_source = propel_data_source.my_data_source.id
  table       = "events"
  timestamp   = "date"
}