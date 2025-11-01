
# 🕸️ Uqda Network

[![Build status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)

## 🌍 Introduction

**Uqda** is an experimental, fully end-to-end encrypted IPv6 mesh network and routing protocol.  
It is **lightweight**, **self-arranging**, and **cross-platform**, enabling any IPv6-capable application to communicate securely with other Uqda nodes — even over the existing IPv4 Internet.

Uqda is designed to be a **future-proof foundation** for decentralized and private networking — a network that anyone can join, extend, or build services upon.

---

## ⚙️ Supported Platforms

Uqda works across multiple platforms:

- 🐧 Linux  
- 🍎 macOS  
- 🪟 Windows  
- 🧱 OpenWrt / VyOS / EdgeRouter  
- 🧊 FreeBSD / OpenBSD  

Additional wrappers and helper scripts are available in the [`contrib`](contrib) directory.

---

## 🧰 Building from Source

To build Uqda manually instead of using pre-built packages:

```bash
# 1. Install Go 1.22+
# 2. Clone the repository
git clone https://github.com/Uqda/Core.git
cd Core

# 3. Build binaries
./build
````

To cross-compile for other systems or architectures:

```bash
GOOS=windows ./build
GOOS=linux GOARCH=mipsle ./build
```

Resulting binaries:

* `uqda` — main daemon
* `uqdactl` — command-line control tool

---

## 🚀 Running Uqda

### Generate Configuration

To generate a configuration file:

**HJSON (human-readable with comments):**

```bash
./uqda -genconf > /path/to/uqda.conf
```

**JSON (machine-friendly):**

```bash
./uqda -genconf -json > /path/to/uqda.conf
```

Then edit `uqda.conf` to add peers, change listen addresses, or tune network behavior.

---

### Start Uqda

**With a static configuration:**

```bash
./uqda -useconffile /path/to/uqda.conf
```

**Auto-configuration mode (ephemeral keys, automatic discovery):**

```bash
./uqda -autoconf
```

> ⚠️ On Linux, you might need to run with `sudo` or grant the binary `CAP_NET_ADMIN` privileges to create TUN/TAP interfaces.

---

## 📚 Documentation

* [CHANGELOG.md](CHANGELOG.md) – version history and updates
* Configuration examples and advanced peering options are available in `/docs` (coming soon)

---

## 🌐 Community & Ecosystem

Community resources will be announced soon — including:

* A public test network
* Peer directories
* Internal DNS & discovery services
* Developer and operator forums

---

## ⚖️ License

This project is released under **LGPLv3**, with an additional exception (from [godeb](https://github.com/niemeyer/godeb)) allowing redistribution of statically or dynamically linked binaries without requiring distribution of Minimal Corresponding Source.

See [LICENSE](LICENSE) for full details.

---

## 💡 Summary

| Feature           | Description                                       |
| ----------------- | ------------------------------------------------- |
| 🔒 Encryption     | Fully end-to-end encrypted IPv6 routing           |
| 🔁 Self-arranging | Automatically discovers and connects peers        |
| 💪 Cross-platform | Works on Linux, macOS, Windows, BSDs, and routers |
| ⚡ Lightweight     | Runs entirely in userspace                        |
| 🌐 Peer-to-peer   | Decentralized by design — no central servers      |

---

## 🧩 Next Steps

If everything builds correctly:

1. Run `go mod tidy` (already done ✅)
2. Build binaries (`./build`)
3. Test your local node:

   ```bash
   ./uqda -autoconf
   ```
4. Join or create a peer network manually by adding peers to `uqda.conf`


