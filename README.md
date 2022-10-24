# Propel Terraform Provider

The [Propel](https://propeldata.com) provider is used to interact with Propel resources, including Data Sources, Data Pools and Metrics. The provider needs to be configured with the proper Application credentials (client ID and secret) before it can be used.

## Requirements
- [Terraform](https://www.terraform.io/downloads.html) 1.2.x
- [Go](https://golang.org/doc/install) 1.17 (to build the provider plugin)

## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-propel
```

## Local release build

```shell
$ go install github.com/goreleaser/goreleaser@latest
```

```shell
$ make release
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install_macos
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```

## Developing the provider

### Running Tests

Configuring tests is similar to configuring the provider. Tests generally assume the following environment variables must be set in order to run tests:
```
PROPEL_CLIENT_ID
PROPEL_CLIENT_SECRET
```

Additional variables may be required for other tests:
```
PROPEL_TEST_SNOWFLAKE_ACCOUNT
PROPEL_TEST_SNOWFLAKE_WAREHOUSE
PROPEL_TEST_SNOWFLAKE_ROLE
PROPEL_TEST_SNOWFLAKE_USERNAME
PROPEL_TEST_SNOWFLAKE_PASSWORD
```

Command to run the acceptance tests:
```
make testacc
```

## Releasing 

Assuming you want to release version x.y.z, you must

1. Update the versions in the following files:

    * Makefile
    * examples/main.tf
    * examples/provider/provider.tf

2. Then, create a tag and push it to main:

    ```
    git tag vx.y.z
    git push origin main vx.y.z
    ```

3. Finally, goreleaser should run and upload artifacts to the GitHub release.
