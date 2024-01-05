# Terraform Provider for Backstage

[![Tests](https://github.com/tdabasinskas/terraform-provider-backstage/actions/workflows/test.yml/badge.svg)](https://github.com/tdabasinskas/terraform-provider-backstage/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/tdabasinskas/terraform-provider-backstage/branch/main/graph/badge.svg?token=1QSZTX0N2B)](https://codecov.io/gh/tdabasinskas/terraform-provider-backstage)
[![go-github release (latest SemVer)](https://img.shields.io/github/v/release/tdabasinskas/terraform-provider-backstage?sort=semver)](https://github.com/tdabasinskas/terraform-provider-backstage/releases)
[![registry](https://img.shields.io/static/v1?label=terraform&message=registry&color=blueviolet)](https://registry.terraform.io/providers/tdabasinskas/backstage/latest)

The [Backstage Provider](https://registry.terraform.io/providers/tdabasinskas/backstage/latest) allows [Terraform](https://terraform.io/) to  manage [Backstage](https://backstage.io) resources.

## Documentation

Official documentation on how to use this provider can be found on the [Terraform Registry](https://registry.terraform.io/providers/tdabasinskas/backstage/latest).
In case of specific questions, please raise a GitHub issue in this repository.

The remainder of this document will focus on the development aspects of the provider.

## Developing

The repository and code is based on [Terraform Provider Scaffolding (Terraform Plugin Framework)](https://github.com/hashicorp/terraform-provider-scaffolding-framework), therefore
most of the official documentation on developing this provider is also applicable.

### Requirements

- [Terraform](https://www.terraform.io/downloads)
- [Go](https://go.dev/doc/install) (1.21)
- [GNU make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/get-docker/) (optional)

### Building

1. `git clone` this repository and `cd` into its directory.
2. `go instal .` to build install the provider into your `$GOPATH/bin` directory.

To be able to run the local version of the provider, please follow the
[official Terraform documentation](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install).

### Testing

In order to test the provider, run the following command:

```bash
make testacc
```

This will run acceptance tests against the provider , actually spawning terraform and the provider, using `https://demo.backstage.io` as the Backstage instance. The instance can
be changed by setting the `BACKSTAGE_BASE_URL` environment variable, e.g.:

```bash
BACKSTAGE_BASE_URL=https://localhost:3000 make testacc
```

The project contains a [`docker-compose.yml`](./docker-compose.yml) file that can be used to spin up a local instance of Backstage, which can be used for testing. To do so, run:

```bash
docker-compose up -d
```

Some tests (i.e. resource tests) do not work with the public [demo instance](https://demo.backstage.io) of Backstage, as they modify the data. To skip those tests while using
the demo instance, run:

```bash
ACCTEST_SKIP_RESOURCE_TEST=1 make testacc
```

### Generating documentation

This provider uses [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs/) to generate documentation and store it in the `docs/` directory.
Once a release is cut, the Terraform Registry will download the documentation from `docs/` and associate it with the release version.
Read more about how this works on the [official page](https://www.terraform.io/registry/providers/docs).

## Releasing

The release process is automated via GitHub Actions, and it's defined in the workflow file [`release.yml`](./.github/workflows/release.yml).

Each release is cut by creating a GitHub release (with corresponding changelog) and pushing a [semantically versioned](https://semver.org/) tag to the default branch.

## Contributing

Contributions to the project are welcome. If you are interested in making a contribution, please review open issues or open a new issue to propose a new feature or bug fix.
Please ensure to follow the code of conduct. Any contributions that align with the project goals and vision are appreciated.
Thank you for your interest in improving the project.

## License

This provider is distributed under the Mozilla Public License v2.0 license found in the [`LICENSE`](./LICENSE) file.
