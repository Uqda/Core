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

## Independent upstream interoperability

```sh
python3 tests/interop/upstream.py
```

Python 3, Go and network access to the pinned upstream module/dependencies are
required. The pin and checksum verification are mandatory; no branch follows
upstream development automatically. The harness builds separate daemon and packet
adapter binaries, prints their hashes and source identities, and checks TCP/TLS
peering, distinct node identities, Uqda address persistence, bidirectional IPv6
frames, relay reconnect and multi-hop forwarding. Every wait/build has a timeout.
See [compatibility](compatibility.md) for the test boundary and source pin.

`go test ./tests/interop` covers the separate same-source daemon harness and
both CLI version commands against the authoritative `src/version/VERSION`.

## Unix privilege-switching tests

`TestCurrentUserid` and `TestCommonUsername` belong to `cmd/uqda`, not a dependency.
They exercise `chuser`, which replaces supplementary groups with the target group
before dropping primary group and user privileges. UID 0 alone is insufficient
on Linux: `setgroups` requires `CAP_SETGID` in the applicable user namespace.
Restricted containers may deny it even to UID 0. Do not suppress such errors in
product code or treat a skipped root-only test as successful execution.

The Linux privilege job compiles the test binary and invokes each test separately
with `sudo`, ensuring the credential change cannot affect another test process.
The runner must have the `nobody` account. To reproduce on a normal Linux host:

```sh
go test -c -o /tmp/uqda-user.test ./cmd/uqda
sudo timeout 30s /tmp/uqda-user.test -test.v -test.run '^TestCurrentUserid$'
sudo timeout 30s /tmp/uqda-user.test -test.v -test.run '^TestCommonUsername$'
```

Compare syscall errors with effective capabilities and a direct `setgroups`
probe when diagnosing an environment restriction. No sandbox-specific skip is used.

## Container validation

`python3 tests/docker/image_test.py` builds the image, checks both version commands,
the `/etc/uqda/uqda.conf` path, identity persistence across containers, migration
and rejection of an invalid legacy identity. It uses a disposable Docker volume,
requires a Docker engine, and does not publish an image or require TUN privileges.

## Race detection and platforms

`go test -race ./...` requires cgo and a supported C compiler. The Linux CI race
job runs this command; a workflow definition alone is not evidence of a passing
run. Windows environments without cgo cannot run the race detector locally.
Mobile Go tests do not replace Android AAR or Apple framework compilation;
see [mobile](mobile.md). Native package installation, upgrade and uninstall tests
must verify that the same public key survives and persistent configuration remains.
