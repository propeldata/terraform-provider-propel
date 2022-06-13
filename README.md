# Propel Terraform Provider

The [Propel](https://propeldata.com) provider is used to interact with resources supported by PropelData. The provider needs to be configured with the proper credentials before it can be used.

## Requeriments
-	[Terraform](https://www.terraform.io/downloads.html) 1.2.x
-	[Go](https://golang.org/doc/install) 1.17 (to build the provider plugin)

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
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```
