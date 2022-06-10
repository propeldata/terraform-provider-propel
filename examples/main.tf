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

resource "propel_datasource" "datasource" {
  unique_name = var.uniqueName
  description = var.description
  connection_settings = {
    username = var.username
    password = var.password
    warehouse = var.warehouse
    role = var.role
    account = var.account
    database = var.database
    schema = var.schema
  }
}

output "datasource" {
  value = propel_datasource.datasource
}
