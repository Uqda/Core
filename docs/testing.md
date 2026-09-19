# Testing

Run from the repository root with Go 1.25 or newer:

```sh
gofmt -w $(git ls-files '*.go')
go build ./...
go vet ./...
go test ./...
govulncheck ./...
```

Install the scanner with `go install golang.org/x/vuln/cmd/govulncheck@latest`.
It requires access to the vulnerability database. Dependency cache contents and
scanner/toolchain versions affect reproducibility; record these in CI results.

## Focused checks

```sh
go test -count=1 ./cmd/uqda ./src/config
go test -count=1 -v ./tests/interop
go test ./src/core -run '^$' -fuzz '^FuzzVersionMetadataDecode$' -fuzztime=10s
go test ./src/multicast -run '^$' -fuzz '^FuzzMulticastAdvertisementUnmarshalBinary$' -fuzztime=10s
go test ./contrib/mobile
go build -o uqda ./cmd/uqda
python3 tests/packaging/config_test.py "$PWD/uqda"
```

The black-box test builds the daemon and runs two local processes, checking
peering and identity-derived addressing. Both processes use the same source.
An independent upstream Yggdrasil fixture is not included. Cross-implementation
validation requires a separately built pinned upstream revision and bidirectional
traffic checks. Do not equate the existing harness with that validation.

## Race detection and platforms

`go test -race ./...` requires cgo and a supported C compiler. The Linux CI race
job runs this command; a workflow definition alone is not evidence of a passing
run. Windows environments without cgo cannot run the race detector locally.
Mobile Go tests do not replace Android AAR or Apple framework compilation;
see [mobile](mobile.md). Native package installation, upgrade and uninstall tests
must verify that the same public key survives and persistent configuration remains.
