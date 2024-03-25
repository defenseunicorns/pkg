# pkg [![Go version](https://img.shields.io/github/go-mod/go-version/defenseunicorns/pkg?filename=helpers/go.mod)](https://go.dev/) [![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/defenseunicorns/pkg/badge)](https://securityscorecards.dev/viewer/?uri=github.com/defenseunicorns/pkg)

## Overview

`pkg` is the monorepo for common Go modules maintained by Defense Unicorns.

## Modules

| Module | Import | Description |
| --- | --- | --- |
| [![GitHub Tag](https://img.shields.io/github/v/tag/defenseunicorns/pkg?sort=date&filter=helpers%2F*&label=docs)](https://pkg.go.dev/github.com/defenseunicorns/pkg/helpers) | `go get github.com/defenseunicorns/pkg/helpers` | Common helper functions for Go. |

## Contributing

1. Install the [pre-commit](https://pre-commit.com/#installation) hooks:

    ```bash
    pre-commit install
    ```

2. Create a new branch on your fork.

    ```bash
    git switch -c <branch>
    ```

3. Make your changes.

4. Run the tests, linters, and formatters.

    ```bash
    make test
    make lint
    make fmt
    ```

5. Commit your changes.

    ```bash
    git commit -m "feat: add new feature"
    ```

6. Push your changes to your fork.

    ```bash
    git push --set-upstream <fork> <branch>
    ```

7. Open a pull request. The title must follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format.

    ```bash
    feat: add new feature
    ```

8. Once your pull request is approved, a new release for any affected modules will be created automatically.

## Testing, Linting, and Formatting

View the [`Makefile`](Makefile) for available commands.

```bash
# Run all formatters
make fmt

# Run all linters
make lint

# Run all tests
make test
```

To run any of the above commands against an individual module, append `-<module name>` to the command.

```bash
# Run all formatters for the helpers module
make fmt-helpers

# Run all linters for the helpers module
make lint-helpers

# Run all tests for the helpers module
make test-helpers
```
