# Baseline (Phase 2)

This is the measured baseline against which all later Uqda refactoring and
optimization must be compared. It records exact commands, exact results, and
— just as importantly — what has **not** been measured yet and why. No
number in this document is estimated or fabricated; anything not actually
run is explicitly marked "not measured."

Baseline commit: `a60ac11` (on top of `87f5d26` docs commit and `27dd82b`
pristine import).

## Environment

| | |
|---|---|
| Machine | Windows 11 Pro, AMD Ryzen 7 5700G, 31.3 GB RAM |
| Go toolchain | `go1.27.0 windows/amd64` |
| `CGO_ENABLED` | `0` (no C compiler installed on this machine — see Limitations) |
| Repo state | Clean working tree at commit `a60ac11` |

## Static checks

### `gofmt -l .`

Found one unformatted file, fixed in commit `a60ac11` (cosmetic-only, see
that commit message): `cmd/genkeys/main.go`. Re-running `gofmt -l .` after
the fix returns no output — tree is fully formatted.

### `go vet ./...`

Clean. No issues reported across any package.

## Build

```
go build ./cmd/uqda ./cmd/uqdactl ./cmd/genkeys
```

Succeeds. All three binaries produced without warnings.

## Unit tests

```
go test ./...
```

All packages pass (this captured output predates the module rename to
`github.com/Uqda/Core` - it's preserved verbatim as an accurate record of
what actually ran at the time, not edited to match the current path):

```
ok    github.com/yggdrasil-network/yggdrasil-go/contrib/mobile   0.134s
ok    github.com/yggdrasil-network/yggdrasil-go/src/address      0.253s
ok    github.com/yggdrasil-network/yggdrasil-go/src/config       0.438s
ok    github.com/yggdrasil-network/yggdrasil-go/src/core         21.349s
ok    github.com/yggdrasil-network/yggdrasil-go/src/ipv6rwc      0.621s
ok    github.com/yggdrasil-network/yggdrasil-go/src/multicast    0.604s
```

(`cmd/*`, `contrib/ansible`, `src/admin`, `src/tun`, `src/version` report "no
test files" — this is the untested-surface gap Phase 16 needs to close, not
a baseline failure.)

## Race detector — BLOCKED on this machine

```
go test -race ./...
```

Fails immediately with `go: -race requires cgo; enable cgo by setting
CGO_ENABLED=1`. This development machine has no C compiler installed
(`gcc` not found, `CGO_ENABLED=0`).

**This is a genuine environment limitation, not a code issue.** Race
detection must run in CI on a Linux runner (where a C toolchain is normally
preinstalled) rather than on this machine.

A `race` job now exists in [.github/workflows/ci.yml](.github/workflows/ci.yml)
(`go test -race ./...` on `ubuntu-latest`), alongside new `vet` and
`vulncheck` jobs, all wired into the `tests-ok` required-checks gate. As of
this commit that workflow is **authored but not yet run**: this checkout
has no git remote configured, so nothing has pushed it to GitHub for
Actions to actually execute. Do not treat "the workflow file exists" as
"race detection passed" until it has actually run at least once — this is
exactly the distinction [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md)
exists to keep honest.

## Dependency vulnerability scan

```
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Result: **0 vulnerabilities reachable from this codebase's own call graph.**
The scan also flags 5 known vulnerabilities in modules that are present in
the dependency graph but not actually called by any code path here:

| ID | Module | Summary |
|---|---|---|
| GO-2026-6355 | `golang.org/x/crypto/ssh` | DoS on deadlocked established channel |
| GO-2026-6354 | `golang.org/x/crypto/ssh` | DoS on deadlocked undecided channel |
| GO-2026-6303 | `golang.org/x/crypto/ssh` | Source-address critical option not enforced for non-public-key auth callbacks |
| GO-2026-5970 | `golang.org/x/text` | Infinite loop on invalid input |
| GO-2026-5932 | `golang.org/x/crypto/openpgp` | Package unmaintained, unsafe by design |

None of these are exploitable through this codebase today (confirmed by
govulncheck's call-graph analysis, not just version matching). They are
transitive dependencies (pulled in by tooling such as the mobile bindings'
toolchain, not by the core networking path). Deliberately **not** bumped as
part of this baseline pass — an unverified dependency version bump is exactly
the kind of change that needs its own test cycle, not a drive-by fix.
Tracked as a Phase 22 (Supply-Chain Security) item.

## Performance — not measured yet

The following are explicitly **not measured** in this baseline, and no
number should be assumed for them until they are:

- Startup time
- Memory at idle
- Memory with multiple peers
- CPU at idle
- Peer establishment time
- Sustained packet forwarding throughput
- Reconnect behavior timing
- Route convergence time
- Teardown/reconnect cycle timing

Measuring these meaningfully requires a running multi-node setup, which is
exactly what the interoperability/lab harness (next step, see
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md)) provides.
Once that harness exists, this document will be updated with real commands
and real numbers — not before.

## Reproducing this baseline

```bash
gofmt -l .
go vet ./...
go build ./cmd/uqda ./cmd/uqdactl ./cmd/genkeys
go test ./...
go test -race ./...   # requires CGO_ENABLED=1 and a C compiler
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```
