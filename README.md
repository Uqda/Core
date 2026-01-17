# 🌐 Uqda Network

<div align="center">

[![Build Status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-LGPL--3.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-00ADD8.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey.svg)](#supported-platforms)

**End-to-End Encrypted • Self-Healing • Zero Configuration**

[Quick Start](#-quick-start) • [Documentation](#-documentation) • [Download](https://github.com/Uqda/Core/releases) • [Community](#-community)

**🇸🇾 [العربية](https://github.com/Uqda/Core/tree/main/docs/ar/README.md)** • **🇬🇧 English**

</div>

---

## 📌 What is Uqda?

**Uqda Network** (from Arabic **عُقدة** meaning "node") is a decentralized routing protocol for building resilient, self-organizing multi-hop mesh networks with end-to-end encryption.

### Key Features

- 🔒 **End-to-End Encrypted** - All traffic encrypted by default (ChaCha20-Poly1305)
- 🌐 **Protocol Compatible** - Works seamlessly with Yggdrasil v0.5 nodes
- ⚡ **Performance Optimized** - 20-50ms latency improvement over baseline
- 🔄 **Self-Healing Mesh** - Automatic path discovery and recovery
- 🎯 **Location Independent** - Permanent IPv6 address derived from your identity
- 🪶 **Zero Configuration** - Networks form automatically
- 💰 **Free Forever** - No cost, no registration, no central authority

---

## ⚡ Quick Start

### Installation

**Linux (Debian/Ubuntu)**
```bash
# Download .deb package from releases
sudo dpkg -i uqda-debian-amd64.deb
```

**Windows**
```powershell
# Download and run the .msi installer
# Or via command line:
msiexec /i uqda-windows-x64.msi
```

**macOS**
```bash
# Download and open the .pkg installer
# Or build from source (see below)
```

**From Source**
```bash
# Prerequisites: Go 1.22+
git clone https://github.com/Uqda/Core.git
cd Core
./build
```

### Running Uqda

**Auto-configuration (Recommended)**
```bash
sudo ./uqda -autoconf
```

**With Configuration File**
```bash
# Generate config
./uqda -genconf > uqda.conf

# Edit uqda.conf as needed, then:
sudo ./uqda -useconffile uqda.conf
```

> **Note:** Root/Administrator privileges are required to create virtual network interfaces.

---

## 🖥️ Supported Platforms

| Platform | Architecture | Package Format |
|----------|--------------|----------------|
| **Linux** | x86_64, ARM64 | `.deb` |
| **Windows** | x86_64, ARM64 | `.msi` |
| **macOS** | Intel, Apple Silicon | `.pkg` |

Download pre-built packages from the [Releases](https://github.com/Uqda/Core/releases) page.

---

## 🏗️ Building from Source

### Prerequisites
- Go 1.22 or later
- Git

### Build Commands
```bash
# Clone repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build for your platform
./build

# Cross-compile examples
GOOS=windows GOARCH=amd64 ./build    # Windows 64-bit
GOOS=linux GOARCH=arm64 ./build      # Linux ARM64
GOOS=darwin GOARCH=arm64 ./build     # macOS Apple Silicon
```

---

## 📚 Documentation

- **[Technical Whitepaper](docs/WHITEPAPER.md)** - Complete technical documentation
- **[Executive Summary](docs/EXECUTIVE_SUMMARY.md)** - One-page overview
- **[FAQ](docs/FAQ.md)** - Frequently asked questions
- **[Security Policy](SECURITY.md)** - Security reporting and best practices
- **[Changelog](CHANGELOG.md)** - Version history and changes
- **[Attribution](ATTRIBUTION.md)** - Credits and licensing information

For detailed documentation, visit the [Wiki](https://github.com/Uqda/Core/wiki).

---

## 🔧 Configuration

Uqda can run in two modes:

### Auto-Configuration Mode
Generates random encryption keys on each startup. Perfect for testing:
```bash
sudo uqda -autoconf
```

### Static Configuration Mode
Uses a persistent configuration file:
```bash
# Generate configuration
uqda -genconf > uqda.conf

# Edit the file to add peers, then:
sudo uqda -useconffile uqda.conf
```

Example peers to add to your config:
```conf
{
  Peers: [
    tcp://[2001:db8::1]:12345,
    tcp://example.com:12345
  ]
}
```

---

## 🤝 Community

- **GitHub Discussions** - [Ask questions & share ideas](https://github.com/Uqda/Core/discussions)
- **Issue Tracker** - [Report bugs](https://github.com/Uqda/Core/issues)
- **Email** - uqda@proton.me

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## 🔒 Security

Found a security vulnerability? Please **do not** open a public issue.

Email us privately at: **uqda@proton.me**

---

## 📄 License

Licensed under **GNU Lesser General Public License v3.0** with binary distribution exception.

```
Copyright (C) 2025-2026 Uqda Network
```

See [LICENSE](LICENSE) for full details.

---

## 🙏 Acknowledgments

Uqda Network is based on the [Yggdrasil Network](https://yggdrasil-network.github.io/) project.

We thank the Yggdrasil team—Neil Alexander, Arceliar, and all contributors—for their pioneering work in decentralized encrypted networking.

For complete attribution, see [ATTRIBUTION.md](ATTRIBUTION.md).

---

<div align="center">

**Made with ❤️ for decentralized networking**

[⬆ Back to Top](#-uqda-network)

</div>