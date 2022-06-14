---
page_title: "Provider: Propel"
subcategory: ""
description: |-
  Terraform provider for interacting with Propel API.
---

# Propel Provider
The [Propel](https://www.propeldata.com) provider is used to interact with the resources supported by Propel. The provider needs to be configured with the proper credentials before it can be used.

## Example Usage

```terraform

terraform {
  required_providers {
    propel = {
      source = "propeldata.com/propeldata/propel"
    }
  }
}

# Configure your Propel provider
provider "propel" {
  client_id = var.propel_client_id
  client_secret = var.propel_client_secret
}
```

## Schema

### Optional

- **client_id** (String) Your Propel Application's clientId value
- **client_secret** (String) Your Propel Application's secret value.