# Propel Terraform Provider

[![Terraform Registry](https://img.shields.io/github/v/release/propeldata/terraform-provider-propel?color=5e4fe3&label=Terraform%20Registry&logo=terraform&sort=semver)](https://registry.terraform.io/providers/propeldata/propel/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/propeldata/terraform-provider-propel)](https://goreportcard.com/report/github.com/propeldata/terraform-provider-propel)

The [Propel](https://www.propeldata.com) Serverless ClickHouse provider interacts with Propel resources, including Data Sources, Data Pools, and Metrics. You need to configure the provider with the proper Application credentials (ID and secret) to get started.

📄 Check out [the documentation](https://registry.terraform.io/providers/propeldata/propel/latest/docs).

🏗 Examples can be found in [examples/](./examples).

❓ Questions? Feel free to create a new issue.

🔧 Want to contribute? Check out [CONTRIBUTING.md](./CONTRIBUTING.md).

## Using the provider

```hcl
terraform {
  required_providers {
    propel = {
      source  = "propeldata/propel"
      version = "~> 1.3.4"
    }
  }
}

variable "propel_application_secret" {
  type = string
  sensitive = true
}

provider "propel" {
  client_id = "APP00000000000000000000000000"
  client_secret = var.propel_application_secret
}
```

We recommend you set your Propel Application's secret via the `TF_VAR_propel_application_secret` environment variable.

# License

This software is distributed under the terms of the MIT license. See [LICENSE](./LICENSE) for details.
