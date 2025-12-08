# 🌐 Uqda Network

<div align="center">

[![Build Status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-LGPL--3.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-00ADD8.svg)](https://golang.org)
[![Platform Support](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20BSD-lightgrey.svg)](#supported-platforms)

**Decentralized • Self-Healing • End-to-End Encrypted**

</div>

---

## 📌 What is Uqda?

**Uqda Network** is an experimental decentralized mesh network protocol that provides:

- 🔒 **End-to-End Encryption** - Fully encrypted IPv6 networking
- 🌍 **Zero Configuration** - Self-arranging mesh topology
- 🔄 **Self-Healing** - Automatic route recovery and optimization
- 🪶 **Lightweight** - Minimal resource footprint
- 🌐 **IPv4/IPv6 Compatible** - Works over existing internet infrastructure
- 🎯 **Location Independent** - Cryptographic addressing that moves with you

> **Note:** Uqda is a fork of the [Yggdrasil Network](https://yggdrasil-network.github.io/) project, rebranded and independently maintained by the Uqda team.

---

## ⚡ Quick Start

### Installation

**Option 1: Install via Go**
```bash
go install github.com/Uqda/Core/cmd/uqda@latest
```

**Option 2: Build from Source**
```bash
# Prerequisites: Go 1.22 or later
git clone https://github.com/Uqda/Core.git
cd Core
./build
```

### Basic Usage

**1. Generate Configuration**
```bash
# Human-friendly HJSON with comments
./uqda -genconf > /etc/uqda/uqda.conf

# Or plain JSON for automation
./uqda -genconf -json > /etc/uqda/uqda.conf
```

**2. Run Uqda**
```bash
# With configuration file
sudo ./uqda -useconffile /etc/uqda/uqda.conf

# Or auto-configuration mode (generates random keys on startup)
sudo ./uqda -autoconf
```

> 💡 **Tip:** On Linux, you can avoid `sudo` by granting `CAP_NET_ADMIN` capability: `sudo setcap cap_net_admin+eip ./uqda`

---

## 🖥️ Supported Platforms

| Platform | Status | Notes |
|----------|--------|-------|
| 🐧 Linux | ✅ Full Support | All major distributions |
| 🍎 macOS | ✅ Full Support | Intel & Apple Silicon |
| 🪟 Windows | ✅ Full Support | Windows 10/11 |
| 🔧 FreeBSD | ✅ Full Support | - |
| 🔧 OpenBSD | ✅ Full Support | - |
| 📡 OpenWrt | ✅ Full Support | Router firmware |
| 🌐 VyOS | ✅ Full Support | Network OS |
| 🔷 EdgeRouter | ✅ Full Support | Ubiquiti devices |

For installation instructions specific to your platform, check the `contrib` folder for platform-specific scripts and tools.

---

## 🏗️ Building from Source

### Prerequisites
- **Go** 1.22 or later ([Download](https://golang.org))
- **Git** for cloning the repository

### Build Steps
```bash
# Clone the repository
git clone https://github.com/Uqda/Core.git
cd Core

# Build for your platform
./build

# Cross-compile for other platforms
GOOS=windows GOARCH=amd64 ./build      # Windows 64-bit
GOOS=linux GOARCH=arm64 ./build        # Linux ARM64
GOOS=darwin GOARCH=arm64 ./build       # macOS Apple Silicon
GOOS=linux GOARCH=mipsle ./build       # MIPS Little-Endian
```

Built binaries will be placed in the project root directory.

---

## 📚 Documentation

**Changelog:** [CHANGELOG.md](CHANGELOG.md) - Version history and updates

**Platform Support:** Check the `contrib` folder for platform-specific scripts and deployment tools.

---

## 🤝 Community

Connect with the Uqda community:

- **GitHub:** [Uqda/Core](https://github.com/Uqda/Core)
- **Email:** [Uqda@proton.me](mailto:Uqda@proton.me)
- **Twitter:** [@tryUqda](https://twitter.com/tryUqda)

We welcome contributions, bug reports, and feature requests!

---

## 🔒 Security

If you discover a security vulnerability, please **do not** open a public issue. Instead, email us directly at [Uqda@proton.me](mailto:Uqda@proton.me) with details.

---

## 📄 License

This project is licensed under the **GNU Lesser General Public License v3.0** with an additional exception for binary distribution.

```
Original work Copyright (C) 2017-2025 Yggdrasil Network
Modified work Copyright (C) 2025 Uqda Network
```

This fork maintains the LGPL-3.0 license of the original Yggdrasil project. Under certain circumstances, this exception permits distribution of binaries that are linked with this code, without requiring the distribution of Minimal Corresponding Source. For details, see [LICENSE](LICENSE).

---

## 🌟 Acknowledgments

Uqda Network is built upon the excellent work of the [Yggdrasil Network](https://yggdrasil-network.github.io/) project. We extend our gratitude to the original developers and contributors.

---

<div align="center">

**Built with ❤️ by the Uqda Team**

[⬆ Back to Top](#-uqda-network)

</div>