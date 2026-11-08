# Development progress log

A running record so work can resume after any interruption. Newest entries
first. Verify state on GitHub before acting: this file can lag behind.

Rules: new work goes to `beta` through pull requests and merges only after CI
passes. `main` moves only by promoting a tested `beta`. Prereleases are tagged
from `beta`; stable releases from `main`.

## Status (2026-11-08)

### Done (merged into beta)

| PR | What |
|---|---|
| #30 | CI fixed for Go 1.27; private-repo CI; CHANGELOG 26.0.3-26.0.5 |
| #31 | Pending-handshake and admin API resource limits |
| #33 | Workflow artifacts expire after 14 days (1,050 old ones deleted) |
| #34 | Stateful inbound firewall, on by default for new nodes |
| #35 | Authenticated release download in CI; Umbrel platform check on demand |
| #36 | Secure invites: `uqda invite` / `uqda join` with QR |
| #37 | Replay-resistant fresh handshake; `RequireFreshHandshake` |
| #38 | `uqda setup` one-step setup with automatic peers |
| #39 | Keep config ownership on rewrite; Windows invite and test fixes |
| #40 | Documentation reorganized; README EN/AR rewritten |

### In progress

| PR | What | State |
|---|---|---|
| #41 | Upstream user requests: password files, `MaxInboundPeers`, log spam, table width, `uqda0` name, key strength; listener leak fix; URI redaction in logs | All CI passes except the CodeQL gate (alert #7, see decisions). Waiting for owner review |
| feat/interface-dns | `IfDNSServers` / `IfDNSDomains` (upstream #1125), settings table in configuration.md | Stacked on #41; local tests pass |

### Next, in order

1. Merge #41 once CI and CodeQL pass.
2. Release `v26.1.0-beta.1` as a prerelease from `beta`: version bump, full
   packaging workflow, assembled release, notes in `docs/releases/`.
3. DNS servers for the Uqda interface (upstream #1125).
4. winget and RPM manifests from the existing MSI and portable builds.
5. Dependabot: Go module group (#29, needs Go 1.26 minimum and docs) and
   Actions group (#32, upload-artifact v7).
6. Umbrel dashboard: firewall rules and invites.
7. `.uqda` names bound to keys (design first).
8. Android app on the existing AAR (design first).

### Decisions that need the owner

- **CodeQL alert #7** (go/clear-text-logging, `src/core/core.go:136`, PR #41).
  The real leak on that line, `?password=` in the logged listener URI, is
  fixed by `RedactURI`. The remaining flow CodeQL reports goes from
  `groupPassword` into `NewPacketConnWithPassword` and then, through the Core
  object, to the listener error. Every error `links.listen` can return
  (network errors, invalid priority, password-file errors with paths only)
  was checked: none contains the group password. Recommended: dismiss as a
  false positive in Security → Code scanning. Dismissing alerts was left to
  the owner on purpose.
- **CodeQL alert #5** (go/disabled-certificate-check, `src/invite/redeem.go`).
  `InsecureSkipVerify` is paired with `VerifyPeerCertificate`, which accepts
  only a certificate whose Ed25519 key equals the key pinned in the invite.
  TLS 1.3 CertificateVerify proves possession of it, and
  `TestImpostorNodeIsRejected` covers an impostor at the same address.
  Recommended: dismiss as a false positive, or as "used in tests" if
  preferred.

- Paid code-signing certificates (Windows Authenticode, Apple notarization).
- App Store and Google Play accounts.
- Final submission of winget and Chocolatey packages from the owner's account.
- When to promote `beta` to `main` and publish a stable 26.1.0.

### Verification notes

- GitHub CI covers Linux, macOS and Windows builds and tests, race detection,
  fuzzing, packaging for every supported system, Docker, mobile SDKs and the
  pinned upstream Yggdrasil interoperability gate.
- Not verified on real hardware: Android and iOS devices, OpenWrt routers,
  physical Umbrel devices.
