# Compatibility

Uqda Core is an independently maintained implementation derived from
[Yggdrasil](https://github.com/yggdrasil-network/yggdrasil-go). It targets the
existing Yggdrasil network and preserves its wire messages, peer negotiation,
Ed25519 identity format and IPv6 address derivation. Ironwood owns routing and
session framing. There is no separate Uqda wire protocol or mandatory extension.

The product version is independent of protocol metadata in `src/core/version.go`.
A product version does not certify compatibility with every upstream release.

## Pinned upstream gate

`tests/interop/upstream.json` pins Yggdrasil **v0.5.14**, commit
`422836eeb21a99790caa286aa63d493bbc4766d7`, together with module and source archive
checksums. `python3 tests/interop/upstream.py` downloads that exact Go module,
extracts it into a separate temporary source directory and builds the unmodified
upstream daemon independently from Uqda. The fixture checks source hashes,
binary module metadata, distinct binary hashes and version output.

The stock daemons peer in both initiating directions over TCP and TLS. The gate
checks identities, address derivation and Uqda restart/reconnect. Separate
packet-adapter processes are built against each implementation's own `core` and
`ipv6rwc` packages. They inject and observe complete IPv6 frames through the
same packet interface used by TUN, with real encrypted sessions and peer links.
They verify bidirectional delivery, relay restart/reconnect and the topology
Uqda A ↔ upstream B ↔ Uqda C, without a direct A–C peering.

The adapter is test infrastructure, not a replacement routing implementation.
This gate requires neither kernel TUN nor network namespaces; it does not certify
OS TUN integration. The Go daemon test in `daemon_test.go` remains a separate
same-source regression test and is not the independent interoperability gate.

## Configuration and identity migration

The canonical binaries are `uqda` and `uqdactl`; the configuration is `uqda.conf`.
Back up the original configuration and any `PrivateKeyPath` target before
migration. Preserve the private key to preserve the network address. Stop the
old node before starting Uqda with that identity; do not run both simultaneously.

Debian, macOS, Windows and container installers recognize their documented
legacy configuration locations. A failed validation must stop migration rather
than create a replacement identity. Review legacy `AdminListen`, `IfName`, log
paths and external key paths manually; copying a configuration does not rewrite
these values. External keys must remain readable to the service account and
must be copied to a protected persistent location before removing upstream files.
Uninstallation of Uqda packages must retain operator-owned configuration and keys.
See [installation](installation.md) and [packaging](packaging.md).

## Mobile API

The exported Go binding type is `mobile.Uqda`. This replaces the upstream
`mobile.Yggdrasil` API and requires client changes and regenerated bindings.
No compatibility alias is provided: an alias would retain the old public brand
and would not reliably preserve generated Java/Objective-C binding names.
See [mobile integration](mobile.md).

Upstream authorship, dependency names, protocol identifiers
and migration paths retain their original names. `contrib/ansible` uses the
external `ansible-yggdrasil` role's variable schema intentionally. The standalone
`contrib/uqda-brute-simple` tool retains its upstream attribution and license.
