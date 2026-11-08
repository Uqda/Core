# Roadmap and gap analysis

Reviewed 2026-11-08 against Core 26.0.5. This is a working plan for
contributors, not a release commitment. Each item states why it matters, how
large it is and whether it can stay wire-compatible with the Yggdrasil network,
which remains a hard requirement ([compatibility](compatibility.md)).

## Where Uqda stands

Uqda already has strong foundations: end-to-end encrypted sessions via Ironwood
(Ed25519 identities, X25519 key agreement, authenticated encryption),
public-key-derived IPv6 addresses, private groups with Argon2id, TCP, TLS,
QUIC, WebSocket, SOCKS and Unix transports, and builds for Linux, Windows,
macOS, FreeBSD, OpenBSD, Debian (six architectures), Docker, Android (AAR),
Apple (XCFramework) and Umbrel. CI covers race detection, fuzzing,
vulnerability scanning, cross-platform CLI tests and real interoperability.

Compared with professional overlay networks (Tailscale, ZeroTier, Nebula), the
gaps are mostly in *usability and control*, not in core encryption.

| Capability | Tailscale / ZeroTier / Nebula | Uqda today |
|---|---|---|
| End-to-end encryption | Yes | Yes (Ironwood) |
| No central coordination server | Nebula partly, others no | **Yes**: fully decentralised |
| Direct connections through NAT (hole punching) | Yes | No: traffic follows peerings, often via relays |
| Human-readable names (MagicDNS-style) | Yes | No: IPv6 addresses only |
| Per-node access rules (who may reach which port) | Yes (ACLs, firewall groups) | No: only `AllowedPublicKeys` (peering) and `GroupPassword` (sessions) |
| Ready-made phone apps | Yes | SDKs only (AAR, XCFramework) |
| Signed installers and verifiable release provenance | Yes | Checksums only; installers unsigned |
| Automatic, verified updates | Yes | No updater |
| Admin API authentication | Yes | None (local-only by design) |
| Metrics for monitoring | Yes | Admin API only |

## Priority 1: security, compatible, small to medium

1. **Make the repository public.** Install instructions, the Linux quick
   installer and the Umbrel app's icon and gallery all point to public GitHub
   URLs that fail while the repository is private. Public visibility also
   enables CodeQL upload, dependency review and artifact attestations at no cost.
2. **Signed release provenance.** Publish Sigstore build attestations
   (`actions/attest-build-provenance`) and an SBOM for every release asset, so
   users can verify that a binary came from this repository's CI rather than
   trusting `SHA256SUMS` hosted next to the files. *Small.* Requires a public
   repository (or GitHub Enterprise).
3. **Replay-resistant handshake extension.** *Done: the fresh handshake (see [security](security.md)).* The current handshake signs only
   the public key (see [security](security.md)). Add an optional second round,
   negotiated through a new metadata field that Yggdrasil nodes ignore, in which
   both sides sign fresh nonces. Add a `RequireFreshHandshake` option for
   restricted peerings so a stripped flag cannot downgrade them. *Medium.* Must
   keep plain Yggdrasil peers working when the option is off.
4. **Fuzz the decrypted IPv6 path.** `src/ipv6rwc` has no fuzz target, unlike
   the handshake and multicast parsers. *Small.*
5. **Optional admin API token** for TCP admin listeners (Unix sockets keep OS
   permissions). *Small.*

## Priority 2: what makes the network useful day to day

6. **Node firewall / access rules.** *Done: see [firewall](firewall.md).* Filter decrypted traffic at `ipv6rwc` by
   remote public key, protocol and port before it reaches TUN, for example
   "only these three keys may reach port 22". This is the most requested
   professional feature, and it is entirely local, so there is no wire change.
   *Medium.*
7. **Names instead of addresses (`.uqda`).** Design agreed, deferred: a
   decentralised `.uqda` namespace with no central registry. Name records are
   signed by the owner's key and carry proof of work against squatting; they
   are stored on many nodes and resolved by asking several and checking
   signatures, with renewal and a warning when a name's key changes. A local
   resolver answers only `.uqda`, alongside personal and group address books.
   Earlier note: A local address book (the Umbrel dashboard
   already has one) served by a small DNS resolver for a reserved suffix such as
   `.uqda.internal`, mapping names to public keys and addresses. *Medium.* Local
   only; sharing name records between nodes is a later step.
8. **Prometheus metrics endpoint** (peers, bytes, sessions, handshake
   rejections), off by default and bound to loopback. *Small.*
9. **Peer quality and automatic selection.** Extend `uqda find` so a new node can
   optionally pick and maintain healthy public TLS/QUIC peers across regions.
   *Medium.*

## Priority 3: reach and polish

10. **Phone apps.** Android `VpnService` and iOS/macOS `NetworkExtension` apps
    built on the existing SDKs, with QR-code peer and group sharing. *Large.*
11. **Signed installers.** Windows Authenticode and macOS notarisation. Needs
    paid certificates. *Small engineering, ongoing cost.*
12. **Verified auto-update** that checks Sigstore provenance before replacing
    binaries, and is opt-in for servers. *Medium.*
13. **Direct paths through NAT.** Research UDP hole punching coordinated over
    the existing overlay, with fallback to current routing. *Large*, and likely
    needs Ironwood changes; coordinate with upstream rather than diverging.
14. **Post-quantum readiness.** Track upstream Ironwood. A hybrid key exchange
    would break wire compatibility, so it cannot ship unilaterally. *Research.*

## Done in this review

- Replay-resistant fresh handshake between Uqda nodes, with `RequireFreshHandshake`.

- One-step onboarding with secure invites ([invites](invites.md)).

- Stateful inbound firewall, on by default for new nodes ([firewall](firewall.md)).

- Bounded pending inbound handshakes (256 per node, 8 per source) and admin API
  resources (64 connections, 1 MiB per request, five-minute idle timeout).
- Fixed CI for Go 1.27 (staticcheck), and for a private repository
  (dependency review, CodeQL upload).
- Completed the CHANGELOG for 26.0.3 to 26.0.5.
