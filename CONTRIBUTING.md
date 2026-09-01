# Contributing to Uqda Core

Thank you for helping improve Uqda. Small, focused changes are the easiest to review.

## Before you start

1. Search existing issues and pull requests.
2. Open an issue before starting a large behavior or protocol change.
3. Never include private keys, live peer addresses, credentials, or personal network configuration in examples or test output.

## Local checks

Uqda Core requires Go 1.25.13 or newer. Before opening a pull request, run:

```bash
gofmt -w .
go vet ./...
go test ./...
go build ./...
```

Add or update tests for behavior changes. Keep platform-specific code behind the appropriate Go build constraints.

## Pull requests

- Explain the problem and the approach used to solve it.
- Link related issues.
- Call out compatibility, protocol, configuration, or security implications.
- Update documentation and `CHANGELOG.md` when user-visible behavior changes.
- Keep generated binaries, local configurations, and editor files out of commits.

By contributing, you agree that your work is licensed under the repository's license.
