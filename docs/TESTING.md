# Testing and release checks

This document lists checks that match (or extend) **CI** (`.github/workflows/ci.yml`) and manual smoke tests before a release.

## Toolchain

CI uses **Go 1.24.x** (see `go.mod`). Run the same minor locally or in CI before tagging a release:

```sh
go version   # expect go1.24.x for parity with CI
```

## Commands (run from repository root)

```sh
go mod verify
go vet ./...
# optional but recommended: fail if any .go file is not gofmt-clean
test -z "$(gofmt -l .)" || { echo "run: gofmt -w on"; gofmt -l .; exit 1; }
go build -v ./...
go test ./... -count=1
go test -race ./... -count=1
```

Cross-compile (same as CI `build-freebsd` job):

```sh
GOOS=freebsd  GOARCH=amd64 go build -v ./...
GOOS=openbsd GOARCH=amd64 go build -v ./...
```

Shell build (embeds version/name via `contrib/semver`):

```sh
./build
./uqda -version
./uqdactl -version
```

**Lint** (config is **golangci-lint v2** — `.golangci.yml`):

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.1 run ./...
```

Or install [golangci-lint v2](https://golangci-lint.run/) and run `golangci-lint run ./...`.

## Optional coverage snapshot

```sh
go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

Coverage is **not** a merge gate in CI today; it helps spot untested packages.

## What is covered automatically

| Area | Automated |
|------|-----------|
| Packages under `go test ./...` | `address`, `config`, `core`, `multicast`, `cmd/uqda` (incl. `chuser` on Unix), `contrib/mobile` |
| `go vet`, `go build`, race detector | All modules in CI matrix (Linux / Windows / macOS) |
| Cross-build | FreeBSD and OpenBSD `GOOS` from Linux runner |
| Static analysis | golangci-lint + CodeQL on GitHub |

## What still needs human / environment testing

These are **not** fully exercised by unit tests alone:

- **TUN / routing** (`src/tun`) — needs root and a real or VM OS (Linux, macOS, Windows, BSD).
- **Admin socket + `uqdactl`** — run `uqda` with a real `uqda.conf`, then `uqdactl getSelf`, `getPeers`, etc.
- **Packaging** — `.deb`, `.msi`, `.pkg`, OpenWrt feeds, VyOS packages: build/install on target images.
- **Long-running mesh** — peering, reconnects, MTU, firewall interaction.

Use [Installation guides](install-linux-manual.md) and the [README](../README.md#documentation) index for per-platform setup.

## Lint `//nolint` directives

Use only **linters that exist in your golangci-lint version** in `//nolint:` comments; unknown names produce warnings (they do not fail the run unless configured to).
