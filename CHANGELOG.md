# Changelog

## Unreleased (beta)

- Inbound overlay firewall. New configurations from `-genconf` and the
  installers drop unsolicited inbound traffic and allow replies to this node's
  own connections, ping and ICMPv6 errors. Rules allow traffic by
  authenticated public key, address or /64, protocol and port. Existing
  configurations keep it off; `uqda status` warns until it is enabled.
  `uqda firewall` and `uqdactl getFirewall` show rules and counters. The
  mobile SDK applies the same configuration. Umbrel keeps it off for now.
  See [firewall](docs/firewall.md).
- Bound pending peer handshakes (256 per node, 8 per source) and admin API
  resources (64 connections, 1 MiB requests, five-minute idle timeout).

## Uqda 26.0.5

Release date: 2026-10-03. Machine version: `26.0.5`; tag: `v26.0.5`.
Full notes: [release-notes-v26.0.5](docs/release-notes-v26.0.5.md).

- Fix the mobile embedding SDK silently ignoring `GroupPassword`. Android and
  Apple bindings now enforce the configured private group; upgrade previously
  distributed bindings before relying on private-group isolation.
- Add the Umbrel service address book and private-group dashboard-access
  workflow in source, with Arabic/English guidance and bounded local-only TCP
  checks. The UI ships through the separately versioned Umbrel wrapper
  (`umbrel-26.0.5-1`).
- Preserve Umbrel's existing login and app gateway; nothing automatically
  exposes apps, changes firewall/routes or bypasses authentication.
- Generalize immutable Umbrel wrapper version validation for the new Core line.

The mobile artifacts are embedding SDKs, not ready-made phone VPN applications.
Installers remain unsigned.

## Uqda 26.0.4

Release date: 2026-10-02. Machine version: `26.0.4`; tag: `v26.0.4`.
Full notes: [release-notes-v26.0.4](docs/release-notes-v26.0.4.md).

- `uqda find` lists read-only public TLS peer suggestions by country from the
  public peers catalog, without contacting peers or reading node configuration.
- `uqda find country=<name> limit=<n>` checks up to eight online TLS candidates
  and ranks reachable TCP ports by connection time, with text and `--json` output.
- Bounded parsing, HTTPS redirect restrictions, stale-snapshot rejection and
  public-IP-only DNS validation protect the discovery path; credential-bearing
  URIs are never suggested.
- Catalog fuzz testing; the two-daemon harness reserves admin and peer ports
  together to avoid duplicate port selection.

Discovery never adds peers or changes `GroupPassword`. A successful TCP check is
not a TLS handshake, key verification or overlay ping.

## Uqda 26.0.3

Release date: 2026-10-02. Machine version: `26.0.3`; tag: `v26.0.3`.
Full notes: [release-notes-v26.0.3](docs/release-notes-v26.0.3.md).

- `uqda help` documents every admin operation, alias, daemon flag, exit code and
  installation lifecycle; `uqda help COMMAND` and `uqdactl help COMMAND` work
  offline even when the daemon is stopped.
- Unix admin-socket permission failures suggest rerunning with `sudo` instead of
  implying daemon failure, return immediately and never echo arguments that may
  contain peer passwords. No automatic elevation or relaxed socket permissions.
- The two-node CLI test harness waits for both daemon startups.

No wire-format or private-group authentication change relative to 26.0.2.

## Uqda 26.0.2

Release date: 2026-10-02. Machine version: `26.0.2`; tag: `v26.0.2`.

- One everyday command center: `uqda` displays node health, with short `peers`,
  `info`, `test`, `version` and `help` commands. Daemon/service flags are unchanged.
- Aligned ASCII status, identity, peer and network-test output for CMD and other
  terminals. Human peer output omits URI credentials and arbitrary error text.
- Reject unknown admin commands/parameters and arguments that were previously
  ignored. Preserve the working advanced diagnostics, available via `uqda commands`.
- Exercise all advertised admin operations on two real private-group nodes in
  a Windows/Linux/macOS CI gate, without touching installed configurations.
- Fix concurrent peer-latency snapshot reads in a pinned local Ironwood snapshot,
  with repeated race regressions. No routing protocol or public-mode changes.
- Harden private-group password derivation with Argon2id, calculated once at
  startup (64 MiB, three passes, four lanes). Private groups must upgrade all
  members together; old private-group authentication is deliberately not accepted.

## Uqda 26.0.1

Released 2026-09-29. Machine version: `26.0.1`; tag: `v26.0.1`.

- Added `uqdactl doctor`/`status` for read-only local node diagnostics.
- Added `uqdactl test <Uqda IPv6>` for bounded bidirectional ICMPv6 reachability
  checks, packet loss and first-probe completion timing, with optional idle
  interval and JSON output. Completion timing includes ping startup, not just RTT.
- Preserved node identity during quick-installer uninstall, reinstall and update
  checks on Ubuntu 24.04 and Fedora 44. The two-node field test transferred
  1 MiB each way with matching SHA-256 values.
- Patch installer versions now increase monotonically after the first general
  release, allowing native package upgrades on Windows and macOS.

Immediately after one Ubuntu service update, 2 of 3 probes were lost while the
peer reconnected; the next 20 all succeeded. Service-active status alone does
not certify instant overlay readiness. Installers remain unsigned and the
administration API must remain private.

## Uqda 26

Released 2026-09-28. Machine version: `26.0.0`; tag: `v26.0.0`.

- First general release of Uqda Core, following the public Beta 1.
- Portable Linux quick installer for systemd on amd64 and arm64, with
  checksum verification, identity-preserving updates/removal, refusal to
  replace unmanaged installations, and explicit identity purge.
- Homebrew Cask for the native macOS packages on Intel and Apple Silicon.
- Ironwood update incorporating upstream stale bloom-bit cleanup and routing
  parent validation, plus opt-in idle-session recovery coverage.
- Cross-platform package builds and native lifecycle checks for Debian,
  Windows MSI, macOS, and portable Linux archives; two-host field tests on
  Ubuntu 24.04 and Fedora 44.

The intermittent upstream idle-session delay was not conclusively reproduced
or fixed. Installers remain unsigned; verify the official `SHA256SUMS` before
installation. The admin API has no authentication, and this release does not
claim anonymity or an independent external security audit.

## Uqda 26 Beta 1

Released 2026-09-27. Machine version: `26.0-beta.1`; tag: `v26.0-beta.1`.

### Highlights

- Uqda Core daemon `uqda`, control utility `uqdactl`, and canonical `uqda.conf`.
- Annual release identity shared by CLI output, build scripts and package metadata.
- Encrypted IPv6 networking using the existing Yggdrasil protocol and Ironwood.

### Security

- Unix warnings for unsafe configuration/key permissions and warnings for
  non-loopback TCP administration sockets.
- Ansible private-key vault files use owner-only permissions.
- Bounds checks and fuzz targets for handshake and multicast metadata parsing.

### Reliability

- Peer removal permits immediate re-addition; reconnect backoff includes jitter.

### Configuration

- `-checkconf` validates keys, peer/listener schemes, admin endpoints and
  multicast interface patterns without starting a node.
- Daemon-loaded configurations require an explicit valid identity; missing or invalid
  keys do not silently generate a replacement node.
- Packaging uses Uqda paths, stops on invalid/ambiguous migration sources, and
  retains persistent configuration on uninstall.

### Compatibility

- Mobile bindings export `mobile.Uqda` and use Uqda product naming.
- Windows MSI uses a distinct Uqda product identity. Configuration migration is
  separate from Windows Installer upgrades; no in-place upstream MSI upgrade is claimed.

### Developer Experience

- Public architecture, security, compatibility, testing and packaging documentation.
- Black-box daemon tests, pinned upstream v0.5.14 interoperability with real IPv6
  packet delivery, TCP/TLS reconnect and multi-hop checks, and Linux race CI.

### Known Limitations

- The admin API has no authentication; protect its socket and keep it local.
- The independent packet gate uses the real IPv6 packet interface without kernel
  TUN; native OS network integration requires separate platform validation.
- Native installers are unsigned beta artifacts; verify the published SHA-256
  checksums before installation.
- No anonymity guarantee or independent external security audit is claimed.

Inherited public upstream release notes are preserved in
[upstream release history](docs/upstream-changelog.md).
