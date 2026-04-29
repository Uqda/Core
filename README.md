# Uqda Core 🌐

[![Build status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)

> **Uqda** (Arabic: عُقَد — nodes/knots) — A fully encrypted, self-organizing IPv6 mesh network. No center. No owner. No single point of failure.

---

## 🤔 How Does It Work?

Imagine the internet, but without any company controlling it.

In a normal network, your traffic goes through servers owned by corporations. In **Uqda**, every device is a node — and nodes talk directly to each other.

```
Normal Internet:          Uqda Network:

You → Company → Friend    You ←——→ Friend
      (they see           (direct, encrypted,
       everything)         no middleman)
```

**Your identity is your key.**
When you join Uqda, a unique cryptographic key is generated for you. Your IPv6 address is mathematically derived from that key — so your address belongs to *you*, not to any registrar or ISP.

```
Your Key → Your Address: 205:xxxx:xxxx:xxxx:...
```

No one can take it away. No one can reassign it.

---

## ✨ What Makes Uqda Different?

| Feature | Traditional VPN | Uqda |
|---------|----------------|------|
| 🏢 Central server | Required | Not needed |
| 🔒 Encryption | Usually yes | Always, end-to-end |
| 📍 Your address | Assigned by provider | Derived from your key |
| 🌐 Routing | Through one server | Across all nodes |
| 💀 Single point of failure | Yes | No |
| 💰 Cost | Often paid | Free & open source |

---

## 🚀 Get Started in 3 Steps

**1 — Download**

Grab the latest release for your platform from the [releases page](https://github.com/Uqda/Core/releases). Each bundle includes both `uqda` and `uqdactl`.

**2 — Generate your config**

```bash
./uqda -genconf > uqda.conf
```

This creates your unique keypair and a ready-to-use config file.

**3 — Connect**

```bash
./uqda -useconffile uqda.conf
```

Your node is now live. Check your address:

```bash
uqdactl getSelf
```

You'll see something like:

```
IPv6 address:  205:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx
IPv6 subnet:   305:xxxx:xxxx:xxxx::/64
Public key:    abc123...
```

---

## 🔗 Connecting to Peers

A peer is any other Uqda node you connect to. The more peers you have, the better your routing across the mesh.

Add peers to your config:

```hjson
Peers: [
  "tls://some-peer.example.com:12345"
  "quic://another-peer.net:9001"
]
```

Or add one at runtime without restarting:

```bash
uqdactl addPeer uri=tls://some-peer.example.com:12345
```

Check who you're connected to:

```bash
uqdactl getPeers
```

---

## 🛠️ Control Your Node

`uqdactl` is your command-line tool for managing a running node.

| Command | What it does |
|---------|-------------|
| `uqdactl getSelf` | Your address, key, and routing info |
| `uqdactl getPeers` | All connected peers and their stats |
| `uqdactl getTree` | The network spanning tree |
| `uqdactl getSessions` | Active encrypted sessions |
| `uqdactl addPeer uri=...` | Add a peer without restarting |
| `uqdactl removePeer uri=...` | Remove a peer |
| `uqdactl list` | All available commands |

---

## 📦 Installation

### Pre-built packages

| Platform | Package |
|----------|---------|
| 🐧 Debian / Ubuntu | `.deb` |
| 🪟 Windows | `.msi` |
| 🍎 macOS | `.pkg` |
| 🔧 EdgeOS / VyOS | `.deb` (router) |

All packages are published automatically on every release.

### Build from source

Requires **Go 1.24+**

```bash
git clone https://github.com/Uqda/Core
cd Core
./build
```

Produces `uqda` and `uqdactl` in the project root.

---

## 📚 Documentation

### Concepts

| Guide | Description |
|-------|-------------|
| [About Uqda](docs/about.md) | How the mesh works, design goals |
| [Configuration](docs/configuration.md) | All config options explained |
| [Configuration reference](docs/configuration-reference.md) | Quick reference table |
| [Advanced peerings](docs/advanced-peerings.md) | Tor, multi-homed, priorities |
| [Private two-node network](docs/private-two-nodes.md) | Closed network between two people |
| [Hosting on Uqda](docs/hosting-on-uqda.md) | Run a website inside the mesh |
| [Mesh DNS](docs/mesh-dns.md) | Set up DNS for your network |
| [Key rotation](docs/key-rotation.md) | What to do if your key leaks |
| [FAQ](docs/faq.md) | Common questions |

### Installation guides

| Platform | Guide |
|----------|-------|
| 🐧 Linux | [Manual build](docs/install-linux-manual.md) |
| 🍎 macOS | [Manual](docs/install-macos-manual.md) · [.pkg installer](docs/install-macos-pkg.md) |
| 🪟 Windows | [Installer](docs/install-windows.md) |
| 📦 Debian / Ubuntu | [Package](docs/install-debian.md) |
| 📦 Fedora / RHEL | [Package](docs/install-rpm.md) |
| 🔧 OpenWrt | [Guide](docs/install-openwrt.md) |
| 🔧 EdgeOS / VyOS | [Guide](docs/install-edgeos.md) |
| 😈 FreeBSD | [Guide](docs/install-freebsd.md) |
| 🐉 Gentoo | [Guide](docs/install-gentoo.md) |

---

## 🔐 Security Notes

- 🔑 **Keep your private key secret** — it is your identity on the network
- 🧱 **Use a firewall** — any node on the mesh can attempt to reach your services
- 🚫 **Uqda does not provide anonymity** — your peers can see your real IP address
- 🔌 **The admin socket has no authentication** — never expose it to untrusted networks

---

## ⚖️ License

Licensed under **GNU LGPLv3** with a library linking exception.
See [`LICENSE`](LICENSE) for the full text.
