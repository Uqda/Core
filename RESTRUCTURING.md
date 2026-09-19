# Restructuring & Versioning Roadmap

**Status: the `cmd/` and packaging rename described below has been executed**
(see [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for the
exact commits). The `src/` subsystem-boundary restructuring (identity/peer/
transport/routing/session/discovery/security/tun/admin/config/diagnostics)
is still in progress - only the identity slice has been done so far. This
document is kept as the target layout and mapping table for the parts not
done yet; entries below that are already complete are marked as such rather
than left implying they're still pending.

It exists so that any contributor or AI agent picking up this work has a
single target to build toward, instead of re-deriving it from chat history.
See [NAMING.md](NAMING.md) for the naming rules this layout follows, and
[ARCHITECTURE.md](ARCHITECTURE.md) for the actual current-state layout.

## Target directory layout

```
Uqda/Core
├── cmd/
│   ├── uqda/            (DONE - was cmd/yggdrasil/)
│   ├── uqdactl/         (DONE - was cmd/yggdrasilctl/)
│   └── genkeys/         (unchanged)
│
├── src/
│   ├── core/            Uqda Core orchestration (from src/core/, minus what moves out below)
│   ├── identity/        Uqda Identity (key handling, currently src/core/core.go + tls.go)
│   ├── address/         Uqda Address (from src/address/, unchanged internally)
│   ├── peer/            Uqda Peer Engine (from src/core/link.go — dial/listen/retry/health)
│   ├── transport/
│   │   ├── tcp/         (from src/core/link_tcp*.go)
│   │   ├── tls/         (from src/core/link_tls.go)
│   │   ├── quic/        (from src/core/link_quic.go)
│   │   ├── websocket/   (from src/core/link_ws.go, link_wss.go)
│   │   ├── socks/       (from src/core/link_socks.go)
│   │   └── unix/        (from src/core/link_unix.go)
│   ├── routing/         Uqda Route — thin wrapper around Ironwood (from src/core/api.go, proto.go)
│   ├── session/         Uqda Session (from src/admin/getsessions.go + any Ironwood-facing session code)
│   ├── discovery/       Uqda Discover (from src/multicast/)
│   ├── security/        Uqda Guard — new subsystem, no direct upstream equivalent
│   ├── tun/             Uqda TUN (from src/tun/ + src/ipv6rwc/)
│   ├── admin/           Uqda Control (from src/admin/, minus getsessions.go which moves to session/)
│   ├── config/          Uqda Config (from src/config/, unchanged internally)
│   ├── diagnostics/     Uqda Doctor — new subsystem, extends uqdactl's existing doctor command
│   └── version/         (from src/version/)
│
├── gateway/
│   ├── deployment/       Uqda Gateway operational profiles
│   ├── monitoring/
│   ├── hardening/
│   └── configs/
│
├── tests/
│   ├── unit/             (existing *_test.go files relocate here or stay colocated — decide during Phase 2 of the audit spec)
│   ├── integration/      (formalizes misc/run-twolink-test, misc/run-schannel-netns as go test-driven)
│   ├── interop/          Uqda ↔ stock Yggdrasil compatibility tests — release-gating
│   ├── fuzz/             new — none exist today
│   ├── chaos/            new — none exist today
│   └── benchmark/        new — none exist today
│
├── contrib/              (unchanged location; content updated for uqda naming over time)
├── docs/
├── third_party/          (attribution for Ironwood, phony, etc. kept separate from Uqda's own NOTICE)
└── .github/
```

## File/directory mapping (current → target)

| Current | Target | Notes |
|---|---|---|
| `cmd/yggdrasil/` | `cmd/uqda/` | **DONE.** Binary renamed `yggdrasil` → `uqda` |
| `cmd/yggdrasilctl/` | `cmd/uqdactl/` | **DONE.** Binary renamed `yggdrasilctl` → `uqdactl` |
| `cmd/genkeys/` | `cmd/genkeys/` | Unchanged |
| `src/core/tls.go` | `src/identity/tls.go` | **DONE** (see IMPLEMENTATION_STATUS.md Phase 3). `src/core/core.go` itself still holds the rest of identity/key state and hasn't moved |
| `src/core/link.go` | `src/peer/` | Generic peer-management logic |
| `src/core/link_tcp*.go`, `link_tls.go`, `link_quic.go`, `link_ws*.go`, `link_socks.go`, `link_unix.go` | `src/transport/{tcp,tls,quic,websocket,socks,unix}/` | Split per-transport |
| `src/core/api.go`, `proto.go` | `src/routing/` | Ironwood-facing wrapper (Ironwood dependency itself is untouched, keeps its own name/license) |
| `src/admin/getsessions.go` | `src/session/` | Session introspection moves out of the generic admin package |
| `src/admin/` (remainder) | `src/admin/` | Conceptually "Uqda Control"; path can stay `admin/` unless a literal rename is desired later |
| `src/multicast/` | `src/discovery/` | |
| `src/tun/`, `src/ipv6rwc/` | `src/tun/` | Merge TUN + the ipv6 read-write-close shim under one subsystem |
| `src/config/` | `src/config/` | **DONE** internally (default config filename/paths are now `uqda.conf` under `/etc/uqda/` etc. - see NAMING.md); package/directory layout itself unchanged |
| `src/version/` | `src/version/` | Unchanged |
| *(none — new)* | `src/security/` | Uqda Guard: connection admission, rate limiting, resource quotas, input validation |
| *(none — new)* | `src/diagnostics/` | Uqda Doctor: expands the existing `uqdactl doctor`-equivalent into a first-class subsystem |
| `contrib/systemd/yggdrasil.service` etc. | `contrib/systemd/uqda.service` etc. | **DONE.** All packaging files (systemd, deb, apparmor, openrc, busybox-init, freebsd, macos, msi, docker, mobile) renamed to match |

## Versioning scheme

Track Uqda Core's own version independently of the Yggdrasil wire-compatibility
version it targets:

```
Uqda Core:                 2.0.0
Yggdrasil wire compatibility: 0.5.x
```

This lets Uqda ship fast (`Uqda Core 3.0`, `4.0`, `5.0`, ...) without ever
implying a new network protocol.

## Phased roadmap (condensed)

The user's full specification defines 25 phases; they are listed here by
number and title only, as a traceability index back to that spec. This
document does not expand phases 4 onward — each gets its own plan when work
on it actually starts.

1. Complete repository audit — **covered by this commit** (ARCHITECTURE.md)
2. Establish a measurable baseline (build/test/perf/memory/CPU/throughput)
3. Refactor Uqda Core into the subsystem boundaries above
4. Build Uqda Guard (security boundary) — **blocked on the hardened-handshake decision in [NAMING.md](NAMING.md#open-question--flag-before-phase-4-uqda-guard-work)**
5. Rebuild peer management (reconnect/backoff/health tracking)
6. Reorganize the transport layer behind a common interface
7. Routing improvements (convergence, recovery) — interop-tested, no wire changes
8. Reliability engineering (abnormal-condition test matrix)
9. Expand Uqda Doctor into a first-class diagnostic tool
10. Build Uqda Gateway operational profiles
11. Document privacy boundaries (no anonymity claims)
12. Identity and key safety hardening
13. Configuration system safety improvements
14. Administration API hardening
15. Logging and metrics standardization
16. Serious automated test suite (unit/integration/interop/fuzz/stress)
17. Uqda Lab (reproducible network test laboratory)
18. Performance engineering, measured against Phase 2's baseline
19. Verifiable builds and releases (SBOM, signing, provenance)
20. Verified update subsystem (does not exist today — net-new)
21. Per-platform support validation and documentation
22. Supply-chain security (dependency pinning, scanning, SBOM)
23. Licensing and attribution preservation (see [ARCHITECTURE.md](ARCHITECTURE.md#legal--attribution-state))
24. Full documentation set
25. Engineering rules for whoever (human or AI) executes the above

## What has and hasn't happened since this document was written

- Binaries: **done** (`cmd/uqda`, `cmd/uqdactl`).
- `go.mod` module path: **done** (`github.com/Uqda/Core`).
- Packaging (systemd/deb/apparmor/openrc/busybox-init/freebsd/macos/msi/docker/mobile): **done**.
- `src/` internal package restructuring (identity/peer/transport/routing/
  session/discovery/security/tun/admin/diagnostics boundaries): **not done**,
  except the identity/TLS slice. This is genuinely large, separate work from
  the product rename and is tracked in
  [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) Phase 3, not
  bundled into the rename commits.
