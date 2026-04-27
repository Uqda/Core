# Uqda Core — documentation hub

This page is the **map of the project**: what the repository contains, how the pieces fit together, and where to read next. For narrative background and design goals, start with **[About Uqda](about.md)**.

---

## Quick start (read order)

1. **[About Uqda](about.md)** — behaviour of the mesh, decentralisation, comparison to VPNs/Tor, project status.  
2. **[Network capabilities & full config guide](network-services-and-config.md)** — what the overlay provides as “services”; every config field in prose; all URI parameters; `uqda` / `uqdactl` tied to config.  
3. **[Configuration](configuration.md)** — config file modes, `genconf`, normalisation, exporting keys.  
4. **[Configuration reference](configuration-reference.md)** — compact table of every `uqda.conf` field and peering URI form.  
5. **[FAQ](faq.md)** — anonymity, ISP visibility, firewall, admin socket, address range, compatibility.  
6. An **install guide** for your OS from the [root README](../README.md#documentation).

---

## What this repository implements

**Uqda Core** is a **user-space IPv6 router** that joins a **cryptographic mesh overlay**:

- Each node has an **Ed25519** identity. A **stable Uqda IPv6 address** (in **`200::/7`**) is derived from the public key, so normal IPv6-aware programs can use the TUN interface without knowing about the mesh.
- **Multi-hop routing** is handled inside the **Ironwood**-based encrypted packet layer (`github.com/Arceliar/ironwood`), wrapped by **`src/core`** as `PacketConn`. Traffic is **end-to-end encrypted** between overlay addresses; underlay links use **TLS** (and other transports) for peering.
- Nodes **peer** over real networks: **static `Peers`**, optional **`Listen`** for inbound, and on LANs **multicast discovery** (`src/multicast`).
- A **TUN** device (`src/tun`, platform-specific files) connects the kernel’s IPv6 stack to the overlay. **`src/ipv6rwc`** handles ICMPv6 and related behaviour on that path.
- **`uqdactl`** talks to the **`src/admin`** socket to introspect peers, paths, sessions, TUN, multicast, and to add/remove peers at runtime.

Nothing here is a hosted “Uqda cloud” — you run **`uqda`** on your own machines.

---

## Programs (`cmd/`)

| Binary | Role |
|--------|------|
| **`uqda`** | Main daemon: loads config, brings up TUN, starts core + multicast + admin. Supports `-genconf`, `-useconffile`, `-autoconf`, `-address` / `-subnet` / `-publickey`, `-normaliseconf`, `-exportkey`, logging, optional user switch on Unix (`cmd/uqda`). |
| **`uqdactl`** | Admin CLI over TCP or Unix socket. Commands include **`list`**, **`getSelf`**, **`getPeers`**, **`getTree`**, **`getPaths`**, **`getSessions`**, **`getNodeInfo`**, **`getMulticastInterfaces`**, **`getTun`**, **`addPeer`**, **`removePeer`**. Endpoint via **`-endpoint`** or inferred from **`AdminListen`** in the default config path. |
| **`genkeys`** | Optional tool to search for keys/addresses with nicer bit patterns (`go run ./cmd/genkeys`); paste chosen keys into config (see [Configuration — stronger addresses](configuration.md#stronger-addresses-and-prefixes)). |

Build all from repo root: **`./build`** (embeds version/name via `contrib/semver/`).

---

## Source layout (`src/`)

| Package | Responsibility |
|---------|----------------|
| **`core`** | Heart of the node: **`Core`** wraps **Ironwood encrypted `PacketConn`**, TLS cert from Ed25519 key, **link** layer (`link_*.go`) for **TCP, TLS, QUIC, WebSocket(s), Unix, SOCKS/sockstls**, listener management, protocol handler, bloom/subnet transform, path notifications, version checks. |
| **`config`** | **`NodeConfig`** struct, HJSON/JSON load, defaults per OS (`defaults_*.go`), key generation and PEM paths. |
| **`address`** | Derives **overlay IPv6** and **`300:`… routed /64** prefix from public key. |
| **`tun`** | Creates and configures the **TUN** interface (Linux, Windows, Darwin, FreeBSD, OpenBSD, …). |
| **`multicast`** | **IPv6 multicast** beacons/listeners on matching interfaces (`MulticastInterfaces` regex, beacon/listen/password/priority). |
| **`admin`** | HTTP-like **admin API** over the admin socket; handlers back `uqdactl` and stats. |
| **`ipv6rwc`** | Read/write control path for IPv6 on TUN (e.g. ICMPv6 handling). |
| **`version`** | Build name/version strings (`-ldflags` from build scripts). |

Reading code: begin at **`cmd/uqda/main.go`**, then **`src/core/core.go`** and **`src/core/link.go`**.

---

## Features (user-visible)

- **Overlay IPv6** with **cryptographic** node identity; **no central address allocator**.
- **Transports between nodes**: **TLS**, **TCP**, **QUIC**, **WebSocket / WSS**, **Unix domain sockets**, **SOCKS / SOCKS+TLS** (see [Configuration reference — Peer URI formats](configuration-reference.md#peer-uri-formats)).
- **URI options**: e.g. **`password`**, **`priority`**, **`maxbackoff`**, **`sni`**, pinned **`key=`** (hex) — see reference table and comments in **`uqda -genconf`** output.
- **Multicast auto-peering** on LAN when enabled; **password-protected** multicast beacons supported.
- **`AllowedPublicKeys`**: restrict **inbound** peering to listed keys (does **not** replace a host firewall; does **not** apply to multicast the same way — see [Private two-node](private-two-nodes.md)).
- **`NodeInfo` / `NodeInfoPrivacy`**: optional metadata visible on the network.
- **`-autoconf`**: ephemeral keys and LAN-oriented defaults for quick tests.
- **Cross-platform** TUN and defaults: Linux, macOS, Windows, FreeBSD, OpenBSD, … (`contrib/` adds packaging, systemd, MSI, EdgeOS, OpenWrt, Docker, AppArmor, etc.).

---

## Configuration at a glance

| Topic | Document |
|--------|-----------|
| All options | **[Configuration reference](configuration-reference.md)** |
| Editing workflow | **[Configuration](configuration.md)** |
| Multi-homed, Tor, priorities, multicast firewall | **[Advanced peerings](advanced-peerings.md)** |
| Exactly two people, no strangers | **[Private network: two people only](private-two-nodes.md)** |
| Compromised key | **[Key rotation after a leak](key-rotation.md)** |
| Remove software completely | **[Uninstall completely](uninstall.md)** |

---

## `contrib/` and packaging (high level)

Not exhaustive; explore `contrib/` in the tree.

| Area | Purpose |
|------|---------|
| **`contrib/semver/`** | Version/name scripts used by **`./build`**. |
| **`contrib/systemd/`, `openrc/`, `launchd/`, …** | Service samples for distros. |
| **`contrib/vyatta-uqda/`** | Ubiquiti EdgeOS / VyOS integration ([install-edgeos](install-edgeos.md)). |
| **`contrib/msi/`, `contrib/macos/`, `contrib/deb/`** | Installer / package scaffolding. |
| **`contrib/mobile/`** | Library build for mobile (see package README / build tags). |
| **`contrib/ansible/`** | Example automation. |
| **`.github/workflows/`** | **CI**: golangci-lint, tests, cross-builds, CodeQL. |

---

## Security and operations (short)

- **Keys**: `PrivateKey` / PEM files are **secret**. **Admin socket** has **no auth** by default — keep Unix socket permissions tight; avoid exposing TCP admin to untrusted networks ([FAQ](faq.md)).  
- **Firewall**: still required for services you expose on Uqda IPv6 ([FAQ](faq.md)).  
- **Release hygiene**: **[Testing and release checks](TESTING.md)** — `go test`, `go vet`, race build, lint, cross-compile.

---

## Protocol and compatibility

Wire protocol carries a **major/minor** version; nodes negotiate compatible paths. **Public mesh** compatibility with other implementations (e.g. Yggdrasil lineage) depends on matching protocol expectations — see [FAQ — Yggdrasil](faq.md).

---

## Glossary

| Term | Meaning |
|------|---------|
| **Overlay** | Virtual IPv6 network built **on top of** TCP/TLS/QUIC/etc. between peers. |
| **Peering** | A **transport session** between two Uqda nodes (not the same as “friend” in social sense). |
| **TUN** | Kernel virtual interface; Uqda reads/writes IPv6 packets there. |
| **Ironwood** | Routing/DHT-style machinery used inside **`core`** for encrypted packet forwarding. |
| **Underlay** | Your normal IP network (ISP, Wi‑Fi, VPN) carrying peer bytes. |

---

## Full documentation index

### Concepts and operation

- [About Uqda](about.md)  
- [Network capabilities & full config guide](network-services-and-config.md)  
- [Hosting a website on Uqda](hosting-on-uqda.md)  
- [Unified DNS for your mesh](mesh-dns.md)  
- [Configuration](configuration.md)  
- [Configuration reference](configuration-reference.md)  
- [Advanced peerings](advanced-peerings.md)  
- [Private network: two people only](private-two-nodes.md)  
- [FAQ](faq.md)  
- [Key rotation after a leak](key-rotation.md)  
- [Uninstall completely](uninstall.md)  
- [Testing and release checks](TESTING.md)  

### Installation

- [Ubiquiti EdgeOS / vyatta-uqda](install-edgeos.md)  
- [Windows](install-windows.md)  
- [macOS (manual build)](install-macos-manual.md)  
- [macOS (.pkg installer)](install-macos-pkg.md)  
- [Linux (manual build)](install-linux-manual.md)  
- [OpenWrt](install-openwrt.md)  
- [Gentoo](install-gentoo.md)  
- [Debian / Ubuntu / Mint](install-debian.md)  
- [RPM (Fedora / RHEL / CentOS)](install-rpm.md)  
- [FreeBSD](install-freebsd.md)  

### In-repo extras

- [Vyatta package README](../contrib/vyatta-uqda/README.md)  
- [License](../LICENSE)  

---

## Licence

The project is released under the terms in **`LICENSE`** in the repository root (LGPLv3 with linking exception as stated there).
