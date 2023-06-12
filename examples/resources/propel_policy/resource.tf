resource "propel_policy" "my_policy" {
  type        = "ALL_ACCESS"
  application = var.my_app.id
  metric      = propel_metric.sum_metric.id
}