terraform {
  required_providers {
    propel = {
      source = "propeldata/propel"
      version = "0.1.1"
    }
  }
}

provider "propel" {
  # Your Propel Application's ID.
  client_id = "APP00000000000000000000000000"

  # Your Propel Application's secret.
  client_secret = var.propel_client_secret
}
