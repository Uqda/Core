# Changelog

## Unreleased (beta)

- Faster recovery when a peer is lost. AutoPeers checks links every 2
  seconds instead of every minute and replaces a lost peer at once from a
  standby list measured in advance, without fetching the catalog. A new
  peer that does not connect within 20 seconds is dropped and the next one
  tried. Extra peers are trimmed by their measured round trip, not only by
  catalog connect time. `uqda autopeers` shows round trips and the standby
  list.
- When the host's network changes (Wi-Fi reconnects, a cable is plugged in,
  an address changes), Uqda redials all peers at once. Before, a peer could
  wait out a reconnect backoff of up to an hour.
- Traffic moves to another path within about 10 ms in most cases when a
  relay's connection closes, instead of stalling for about 2.7 s: a broken
  path is looked up again after 250 ms instead of 1 s. When the lost link
  was the node's parent in the routing tree, or the link goes silent,
  recovery still takes about 3 to 4 seconds, bounded by the routing
  protocol.
- CI measures failover on every change: four nodes, two paths, the path in
  use cut by closing it and by silently dropping packets.
- Windows: the installer puts the Uqda folder on the system `PATH`, so
  `uqda status`, `uqda join` and the other commands work in a new terminal.
  Uninstalling removes it again.
- Invites list at most four addresses, real LAN addresses first, and skip
  virtual adapters (WSL, Hyper-V, Docker, VPN tunnels), so joining is faster
  and the QR code is smaller and easier to scan.
- Windows: the installer adds a firewall rule for `uqda.exe` limited to the
  local subnet on private networks, so local discovery, LAN peerings and
  invites are not silently blocked. Removed on uninstall.
- `uqda info` shows the release name ("Uqda Core 26.1") like the other
  commands.
- Privacy: new configurations set `NodeInfoPrivacy: true`, so a node no
  longer tells others its platform and Uqda version unless you choose to.
  Existing configurations are unchanged. A new [privacy](PRIVACY.md) page
  states what a node shares, with whom, and that Uqda contains no telemetry.
- Licenses: [THIRD_PARTY_LICENSES](THIRD_PARTY_LICENSES.md) lists every
  component in the release binaries, and CI fails when a new dependency is
  missing from it or has a license outside the allowed set. The `.deb`, MSI
  and Docker image now install `LICENSE`, `NOTICE.md`, the third-party list
  and `PRIVACY.md`; the MSI also ships the Wintun driver license.
- [SECURITY](SECURITY.md) adds an e-mail contact, response targets, a safe
  harbour for good-faith research and an abuse section. NOTICE records the
  fork point and Uqda's modifications. README states that Uqda is not
  anonymous and not affiliated with the Yggdrasil project, in English and
  Arabic. [Legal notes](docs/legal/README.md) list what is done and what
  remains for the maintainers.

## Uqda 26.1

Release date: 2026-11-08. Machine version: `26.1.0`; tag: `v26.1.0`.
Full notes: [Uqda 26.1](docs/releases/v26.1.0.md). Uqda is an experimental project.

Everything in 26.1 Beta 1 below, plus:

- Automatic peer selection. `AutoPeers: { Enabled: true, Country: "germany",
  Count: 2 }` keeps that many healthy outbound peers: Uqda adds the fastest
  reachable public TLS peers, replaces one it added after five minutes down,
  and swaps one for a much faster peer on each check (every 30 minutes by
  default). Configured `Peers` are never removed. `uqda setup country=NAME`
  turns it on; `uqdactl getAutoPeers` and `uqda status` report it. See
  [public peers](docs/public-peers.md#automatic-peers).

- The Debian package installs in containers and chroots: its scripts start
  or stop the service only when systemd is running, and otherwise print how
  to start Uqda instead of failing with a half-configured package.
- The configuration permission warning no longer fires for the packaged
  `root:uqda 0640` layout. It warns when other users have any access or the
  group can write or execute the file.

## Uqda 26.1 Beta 1

Prerelease date: 2026-11-08. Machine version: `26.1-beta.1`; tag: `v26.1-beta.1`.
Full notes: [26.1 Beta 1](docs/releases/v26.1-beta.1.md). Uqda is an experimental project.

- `uqda setup [mode=private|public] [country=NAME] [peers=0..5]` sets up a node
  in one step. On a new device it creates an identity, a private network with
  a random 192-bit group password, the inbound firewall, and the fastest
  healthy public TLS peers in the chosen country. On an existing node it keeps
  the identity and adds only what is missing. Switching a private node to
  public needs `confirm=yes`; a backup is kept as `PATH.before-setup`.

- Fresh handshake. Uqda nodes add a random nonce to the peer handshake and,
  when both sides support it, each signs the other's nonce before the link
  starts, so a recorded handshake can no longer be replayed to impersonate a
  peer or pass `AllowedPublicKeys`. Yggdrasil peers are unaffected. The new
  `RequireFreshHandshake` option rejects peers without it, for all-Uqda
  networks. `getPeers` reports `fresh_handshake`.

- Invites. `uqda invite [expires=15m] [uses=1] [name=TEXT]` prints a one-time,
  expiring invite as text and a QR code. `uqda join INVITE` contacts the
  inviting node over TLS 1.3 pinned to its public key, proves possession of
  the invite's random secret, and receives the group password and peers,
  keeping its own identity. Invites hold no password, are rate-limited and
  vanish on restart. `uqdactl cancelInvites` revokes them; `InvitePort` fixes
  the redemption port. See [invites](docs/invites.md).

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
Full notes: [release-notes-v26.0.5](docs/releases/v26.0.5.md).

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
Full notes: [release-notes-v26.0.4](docs/releases/v26.0.4.md).

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
Full notes: [release-notes-v26.0.3](docs/releases/v26.0.3.md).

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
[upstream release history](docs/upstream/changelog.md).
