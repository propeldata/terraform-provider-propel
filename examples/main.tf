terraform {
  required_providers {
    propel = {
      version = ">= 0.0.1"
      source = "propeldata.com/propeldata/propel"
    }
  }
}

provider "propel" {
  client_id = var.client_id
  client_secret = var.client_secret
}

resource "propel_datasource" "test-datasource" {
  unique_name = var.uniqueName
  description = var.description
  username = var.username
  password = var.password
  warehouse = var.warehouse
  role = var.role
  account = var.account
  database = var.database
  schema = var.schema
}

output "test-datasource" {
  value = propel_datasource.test-datasource.id
}
