# pkg [![Go version](https://img.shields.io/github/go-mod/go-version/defenseunicorns/pkg?filename=helpers/go.mod)](https://go.dev/) [![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/defenseunicorns/pkg/badge)](https://securityscorecards.dev/viewer/?uri=github.com/defenseunicorns/pkg)

## Overview

`pkg` is the monorepo for common Go modules maintained by Defense Unicorns.

## Modules

| Module | Import | Description |
| --- | --- | --- |
| [![GitHub Tag](https://img.shields.io/github/v/tag/defenseunicorns/pkg?sort=date&filter=exec%2F*&label)](https://pkg.go.dev/github.com/defenseunicorns/pkg/exec) | <pre lang="bash">go get -u github.com/defenseunicorns/pkg/exec</pre> | provides a wrapper around the os/exec package |
| [![GitHub Tag](https://img.shields.io/github/v/tag/defenseunicorns/pkg?sort=date&filter=helpers%2F*&label)](https://pkg.go.dev/github.com/defenseunicorns/pkg/helpers/v2) | <pre lang="bash">go get -u github.com/defenseunicorns/pkg/helpers/v2</pre> | provides common helper functions for Go. |
| [![GitHub Tag](https://img.shields.io/github/v/tag/defenseunicorns/pkg?sort=date&filter=kubernetes%2F*&label)](https://pkg.go.dev/github.com/defenseunicorns/pkg/kubernetes) | <pre lang="bash">go get -u github.com/defenseunicorns/pkg/kubernetes</pre> | provides helper functions related to Kubernetes. |
| [![GitHub Tag](https://img.shields.io/github/v/tag/defenseunicorns/pkg?sort=date&filter=oci%2F*&label)](https://pkg.go.dev/github.com/defenseunicorns/pkg/oci) | <pre lang="bash">go get -u github.com/defenseunicorns/pkg/oci</pre> | provides tools for interacting with artifacts stored in OCI registries |

## Contributing

Follow the steps in [CONTRIBUTING.md](./.github/CONTRIBUTING.md) to contribute to this project.

## Testing, Linting, and Formatting

View the [`Makefile`](Makefile) for available targets.

```bash
# Run all formatters
make fmt

# Run all linters
make lint

# Run all tests
make test
```

To run any of the above against an individual module, append `-<module name>` to the target.

```bash
# Run all formatters for the helpers module
make fmt-helpers

# Run all linters for the helpers module
make lint-helpers

# Run all tests for the helpers module
make test-helpers
```
