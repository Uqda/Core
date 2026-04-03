# Changelog

All notable changes to **this repository** are recorded here.

This project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html) and [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

**Lineage note:** Git history on `main` was reset to start at **1.0.0** as a single baseline that contains the full current tree. Earlier per-commit history is not preserved on `main`. The codebase descends from the [Yggdrasil Network](https://github.com/yggdrasil-network/yggdrasil-go) project; see [ATTRIBUTION.md](ATTRIBUTION.md). For upstream Yggdrasil release notes prior to the Uqda fork, see [yggdrasil-go releases](https://github.com/yggdrasil-network/yggdrasil-go/releases).

---

## [1.0.0] - 2026-04-04

First unified public release of **Uqda Core** as version **1.0.0**. This tag represents the complete product as currently shipped: mesh node, control tooling, documentation, packaging, and CI.

### Vision & scope

End-to-end encrypted IPv6 overlay mesh (`200::/7`), self-organizing routes, stable addresses derived from Ed25519 keys, multi-transport peerings, and operational controls (admin socket, optional Web UI, metrics). See [PROJECT_VISION.md](PROJECT_VISION.md) for the full narrative (Arabic + English).

### Highlights

- **Node (`uqda`)** — IPv6 overlay, ironwood-derived routing, TUN/TAP, multicast discovery, multiple link types.
- **Control (`uqdactl`)** — Admin socket client; network/peer workflows where implemented.
- **Security & ops** — `AllowedPublicKeys`, private network / invite flows, admin auth token support, security policy and vulnerability scanning in CI.
- **Transports** — TCP, TLS, QUIC, WebSocket (incl. WSS), SOCKS, UNIX; IPv4 underlay DNS options (`PeerDialNetwork`, `PreferIPv4`) while the overlay remains IPv6.
- **Observability** — Optional Prometheus `MetricsListen`, optional Web UI (`UIListen`) with Admin API proxy.
- **Toolchain** — Go **1.25+** with **toolchain go1.25.8** in `go.mod`; CI runs tests, `govulncheck`, and golangci-lint.

### Attribution

Uqda is a fork and rebrand of **Yggdrasil Network**; credit and license obligations are documented in [ATTRIBUTION.md](ATTRIBUTION.md).

---

## Earlier numbering (informational only)

Packages and docs may still mention **0.1.x** in passing; treat **1.0.0** as the canonical first-line release ID for this repository going forward.
