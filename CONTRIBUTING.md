# Contributing

All contributions are welcome, whether they are technical in nature or not.

Feel free to open a new issue to ask questions, discuss issues or propose enhancements.

The rest of this document describes how to get started developing on this repository.

## What should I know before I get started?

### Terraform documentation

Hashicorp has a lot of documentation on creating custom Terraform providers categorized under [Extending Terraform](https://www.terraform.io/docs/extend/index.html). This might help when getting started, but are not a pre-requisite to contribute. Feel free to just open an issue and we can guide you along the way.

### Propel documentation

Our Terraform provider is used for exposing and configuring Propel resources. Everything it does is made possible by Propel's GraphQL API. We have documented what each Propel resource does and how the GraphQL API works in [Propel's documentation](https://www.propeldata.com/docs).

## Contributing changes

### Preview document changes

Hashicorp has a tool to preview documentation. Visit [registry.terraform.io/tools/doc-preview](https://registry.terraform.io/tools/doc-preview).

### Running the tests

Most of the tests are acceptance tests, which will call real APIs. To run the tests you'll need to have access to a Propel account.

First, **create an Application** within your Propel account. Ensure you grant "admin" scope to the Application. Keep track of your Application's ID and secret.

Next, **run the acceptance tests** by passing the Application ID and secret as environment variables:

```sh
PROPEL_CLIENT_ID=<your Application ID> PROPEL_CLIENT_SECRET=<your Application secret> make testacc
```

### Using a locally built version of the provider

It can be handy to run `terraform` with a local version of the Propel provider.

First, **build the provider:**

```
make
```

Next, **install the provider**. The provider has to be installed in one of the [local mirror directories](https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories) using the [new filesysem structure](https://www.terraform.io/upgrade-guides/0-13.html#new-filesystem-layout-for-local-copies-of-providers). If you are on macOS, you can do this as follows:

```
make install_macos
```

Other operating systems should be similar, so feel free to add an additional target.

### Enabling log output

To print logs (including full dumps of requests and their responses), you have to set `TF_LOG` to at least `debug` when running `terraform`:

```sh
TF_LOG=debug terraform apply
```

For more information, see [Debugging Terraform](https://www.terraform.io/docs/internals/debugging.html).

### Style convention

CI will run the following tools to style code:

```sh
goimports -l -w .
go mod tidy
```

`goimports` will format the code like `gofmt`, but it will also fix imports. It can be installed with

```
go get golang.org/x/tools/cmd/goimports
```

Both commands should create no changes before a pull request can be merged.

### Release Procedure

To release a new version of the Terraform provider a binary has to be built for a list of platforms ([more information](https://www.terraform.io/docs/registry/providers/publishing.html#creating-a-github-release)). This process is automated with GoReleaser and GitHub Actions.

1. Update the versions in the following files to the version you intend to release:

    * Makefile
    * examples/main.tf
    * examples/provider/provider.tf
2. Then re-generate the documentation by running the command:

   ```shell
   go generate
   ```

3. Create [a new release](https://github.com/propeldata/terraform-provider-propel/releases/new).
4. The tag and release title should be a semantic version.
5. To follow the convention of other Terraform providers, the description should have the following sections (each section can be omitted if empty):

    ```text
    NOTES:
    FEATURES:
    ENHANCEMENTS:
    BUG FIXES:
    ```

6. After that tag has been created, a GitHub Actions workflow will run and add binaries to the release.
7. Once the tag is created, the [Terraform Registry](https://registry.terraform.io/providers/propeldata/propel/latest) should also list the new version.
