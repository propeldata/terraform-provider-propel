terraform {
  required_providers {
    hashicups = {
      version = ">= 0.3.1"
      source = "hashicorp.com/edu/hashicups"
    }
  }
}

provider "propel" {
  client_id = var.client_id
  secret = var.secret
}

resource "datasource" "test-datasource" {
  uniqueName = var.uniqueName
  description = var.description
  username = var.username
  password = var.password
  warehouse = var.warehouse
  role = var.role
  account = var.account
}

output "psl" {
  value = datasource.test-datasource.id
}
