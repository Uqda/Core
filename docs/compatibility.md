# Compatibility

Uqda Core is an independently maintained implementation derived from
[Yggdrasil](https://github.com/yggdrasil-network/yggdrasil-go). It targets the
existing Yggdrasil network and preserves its wire messages, peer negotiation,
Ed25519 identity format and IPv6 address derivation. Ironwood owns routing and
session framing. There is no separate Uqda wire protocol or mandatory extension.

The product version is independent of protocol metadata in `src/core/version.go`.
A product version does not certify interoperability with a particular upstream
release. The repository's two-process daemon test uses the same Uqda source for
both peers. No independently built, pinned upstream fixture is included, so it
is not proof of Uqda-to-upstream interoperability.

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

Upstream authorship, dependency names, protocol identifiers, third-party tools
and migration paths retain their original names. `contrib/ansible` uses the
external `ansible-yggdrasil` role's variable schema intentionally. The standalone
`contrib/yggdrasil-brute-simple` tool and its license retain their upstream names.
