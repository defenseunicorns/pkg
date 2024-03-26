# Contributing to `defenseunicorns/pkg`

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

    > Set up your Git config to GPG sign all commits. [Here's some documentation on how to set it up](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits). You won't be able to merge your PR if you have any unverified commits.

    ```bash
    git commit -m "feat: add new feature"
    ```

6. Push your changes to your fork.

    ```bash
    git push --set-upstream <fork> <branch>
    ```

7. Open a pull request. The title must follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format (`fix:`, `feat:`, `chore:`, `docs:`, etc...). For example:

    ```bash
    feat: add new feature
    ```

    > Use `wip:` if you are unsure about the changes and want feedback about the scope of the PR.

8. Once your pull request is approved, a new release for any affected modules will be created automatically.
