---
page_title: "Provider: Propel"
subcategory: ""
description: |-
  Terraform provider for interacting with Propel API.
---

# Propel Provider
The [Propel](https://propeldata.com) provider is used to interact with Propel resources, including Data Sources, Data Pools and Metrics. The provider needs to be configured with the proper Application credentials (client ID and secret) before it can be used.

## Example Usage

```terraform

terraform {
  required_providers {
    propel = {
      source = "propeldata.com/propeldata/propel"
    }
  }
}

# Configure the provider to use your Propel Application
provider "propel" {
  client_id = var.propel_client_id
  client_secret = var.propel_client_secret
}
```

## Schema

### Optional

- **client_id** (String) Your Propel Application's client ID
- **client_secret** (String) Your Propel Application's client secret