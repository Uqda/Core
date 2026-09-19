# Uqda Implementation Status

Living document. Updated every time a workstream changes state. Status values:
**Not started**, **In progress**, **Implemented**, **Tested**, **Blocked**, **Deferred**.

A workstream is only marked **Tested** when there is an actual test, benchmark,
or verification command backing the claim — not because code exists. See
linked evidence in each row.

Last updated: after commit `48dd761` (admin socket exposure warning).

## Phase tracker

| # | Phase | Status | Evidence / notes |
|---|---|---|---|
| 1 | Repository audit | **Implemented** | [ARCHITECTURE.md](../ARCHITECTURE.md), commit `87f5d26` |
| 1 | Naming specification | **Implemented** | [NAMING.md](../NAMING.md), commit `87f5d26` |
| 1 | Restructuring roadmap | **Implemented** | [RESTRUCTURING.md](../RESTRUCTURING.md), commit `87f5d26` |
| 2 | Baseline: format/vet/build/test | **Tested** | [BASELINE.md](../BASELINE.md), commits `a60ac11` — gofmt clean, vet clean, build clean, `go test ./...` all green |
| 2 | Baseline: race detection | **Blocked** | No cgo/C compiler on the current dev machine; must run in CI on Linux. See BASELINE.md "Race detector" section |
| 2 | Baseline: dependency vulnerability scan | **Tested** | `govulncheck` run, 0 reachable vulnerabilities; 5 unreachable transitive ones logged in BASELINE.md for Phase 22 |
| 2 | Baseline: performance numbers (startup/memory/CPU/throughput/convergence) | **Not started** | Requires the interop/lab harness below to exist first — see BASELINE.md "Performance — not measured yet" |
| 3 | Structural refactor (identity/peer/transport/routing/session/discovery/security/tun/admin/config/diagnostics boundaries) | **Not started** | Target layout defined in RESTRUCTURING.md; not yet executed |
| 4 | Uqda Guard (security boundary) | **Not started** | Blocked on the hardened-handshake decision — **now resolved**: no new wire messages, implementation-side hardening only (see NAMING.md "Open question", now closed) |
| 5 | Peer lifecycle rebuild | **Not started** | |
| 6 | Transport layer reorganization | **Not started** | |
| 7 | Routing improvements | **Not started** | High-risk; requires topology simulation before any change, per policy |
| 8 | Reliability engineering (abnormal-condition matrix) | **Not started** | |
| 9 | Uqda Doctor expansion | **Not started** | `uqdactl` currently has no `doctor` subcommand in this checkout — to confirm during Phase 9 scoping |
| 10 | Uqda Gateway deployment profile | **Not started** | |
| 11 | Privacy boundary documentation | **Not started** | |
| 12 | Identity/key safety hardening | **In progress** | [contrib/ansible/genkeys.go](../contrib/ansible/genkeys.go) vault file now created at 0600 (commit `4a8b113`, tested). Daemon now warns on unsafe config/`PrivateKeyPath` file permissions on Unix, no-ops correctly on Windows ([src/config/permissions.go](../src/config/permissions.go), commit `4bf1891`, tested). Not yet done: corrupted/missing-identity handling audit, upgrade-preservation tests |
| 13 | Configuration safety improvements | **Not started** | `-checkconf`-equivalent validation command not yet implemented |
| 14 | Admin API hardening | **In progress** | [src/admin/admin.go](../src/admin/admin.go): warns when the admin socket binds to a non-loopback TCP address, explaining the lack of authentication (commit `48dd761`, tested: loopback/0.0.0.0/unix-socket cases). Actual authentication on the admin socket itself is still not implemented — the TODO in that file's header is still accurate |
| 15 | Logging/metrics standardization | **Not started** | |
| 16 | Full automated test suite (unit/integration/interop/fuzz/stress) | **In progress** | [tests/interop/daemon_test.go](../tests/interop/daemon_test.go): black-box two-process daemon test (config file + admin socket, not the internal API) verifying peering and address-derivation correctness. Currently same-codebase only — becomes a true cross-implementation test once `cmd/uqda` exists (see that file's package doc). Fuzz: `FuzzVersionMetadataDecode` and `FuzzMulticastAdvertisementUnmarshalBinary` added, both run 2.5M+/3.1M+ execs with zero crashes. Stress/chaos still not started |
| 17 | Uqda Lab (network laboratory) | **Not started** | |
| 18 | Performance engineering | **Not started** | Depends on Phase 2 performance baseline |
| 19 | Verifiable builds/releases (SBOM, signing) | **Not started** | No git remote/CI configured yet on this checkout |
| 20 | Verified update subsystem | **Not started** | Net-new; no update mechanism exists upstream |
| 21 | Platform support validation | **Not started** | |
| 22 | Supply-chain security | **In progress** | Initial `govulncheck` pass done (see Phase 2 row); no dependency bumps made yet |
| 23 | Licensing/attribution (NOTICE.md) | **Not started** | |
| 23 | Threat model | **Implemented** | [docs/THREAT_MODEL.md](THREAT_MODEL.md): 10 concrete threats mapped to actual file/line boundaries in this codebase. 3 of the identified gaps have since been addressed (admin socket exposure now warns, config/key file permissions now warn, ansible vault file now 0600) and the doc was updated in place with the actual fix + test reference for each rather than left stale. One entry (post-handshake connection deadlines) was corrected after checking Ironwood's source directly — it already enforces a 3s peer timeout with 1s keepalives, which the original entry had incorrectly called unmitigated |
| 24 | Full documentation set | **In progress** | ARCHITECTURE.md, NAMING.md, RESTRUCTURING.md, BASELINE.md, [SECURITY.md](../SECURITY.md), [docs/THREAT_MODEL.md](THREAT_MODEL.md) exist; README rewrite, Arabic README, upstream-comparison doc still pending |
| 25 | Engineering rules | **Implemented** | Encoded in this repo's working process (small coherent commits, test-before-refactor, no wire changes) rather than as a separate document |

## Blockers requiring external input

- **Race detection on the primary dev machine**: needs a C toolchain
  (e.g. installing a MinGW/TDM-GCC toolchain) or must be deferred entirely
  to a Linux CI runner. Not resolved autonomously since installing new
  system-level dev tooling wasn't part of the original scope — flagging for
  a decision rather than silently installing a compiler toolchain.
- **CI/release infrastructure (Phases 19, 21, 22 automation)**: this
  checkout has no git remote configured, so GitHub Actions workflows can be
  authored but cannot actually run until the repo is pushed somewhere.
- **Code signing / package repository credentials (Phase 19, 21)**: requires
  externally-provisioned certificates/accounts not available in this
  environment — explicitly out of scope per the "no external credentials"
  boundary.

## Hardened-handshake decision (resolved)

Per the non-negotiable compatibility rule in [NAMING.md](../NAMING.md): no
Uqda-specific wire handshake or mandatory extension will be introduced.
Uqda Guard (Phase 4) will be implemented entirely through
implementation-side controls — parsing safety, validation, resource/queue
limits, timeouts, rate limiting — none of which add anything a stock
Yggdrasil peer wouldn't already tolerate on the wire.
