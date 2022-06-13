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
  unique_name = var.datasource_unique_name
  description = var.datasource_description
  connection_settings = {
    username = var.datasource_username
    password = var.datasource_password
    warehouse = var.datasource_warehouse
    role = var.datasource_role
    account = var.datasource_account
    database = var.datasource_database
    schema = var.datasource_schema
  }
}

resource "propel_datapool" "datapool" {
  unique_name = var.datapool_unique_name
  description = var.datapool_description
  datasource = propel_datasource.datasource.id
  table = var.datapool_table
  timestamp = var.datapool_timestamp
}

output "datasource" {
  value = propel_datasource.datasource
}

output "datapool" {
  value = propel_datapool.datapool
}
