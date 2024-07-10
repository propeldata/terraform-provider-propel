resource "propel_application" "my_application" {
    unique_name = "My Application"
    description = "This is an example of an Application"
    propeller = "P1_LARGE"
    scopes = ["METRIC_QUERY", "DATA_POOL_QUERY"]
}