# Uqda Core

[![Build status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)

## About

Uqda Core provides an implementation of a fully end-to-end encrypted IPv6 network.
See [https://github.com/Uqda/Core](https://github.com/Uqda/Core) for documentation and updates.

## Installation

See the [releases page](https://github.com/Uqda/Core/releases) for pre-built binaries.

**Maintainers — automatic releases:** pushing a **semver tag** matching `v1.2.3` runs [`.github/workflows/release.yml`](.github/workflows/release.yml), which builds Linux/Windows/macOS artifacts and publishes **one bundle per platform** (each bundle contains both `uqda` and `uqdactl`) plus `SHA256SUMS` on the GitHub Release. Example: `git tag -a v0.1.0 -m "Uqda Core 0.1.0" && git push origin v0.1.0`.

**Installers on the same release page:** [`.github/workflows/pkg.yml`](.github/workflows/pkg.yml) now also runs on `v*` tags and uploads `.deb` / `.pkg` / `.msi` (plus router packages and vendored sources) to the same GitHub Release, so users can install via one platform package without extra manual assembly.

## Documentation

**Start here:** [Documentation hub — full project map](docs/README.md) (architecture, packages, features, links to every guide).

### Concepts and operation

- [About Uqda](docs/about.md)
- [Network capabilities & full config guide](docs/network-services-and-config.md) — what the mesh provides; every `uqda.conf` field; URI options; `uqda` / `uqdactl` surfaces
- [Hosting a website on Uqda](docs/hosting-on-uqda.md)
- [Unified DNS for your mesh](docs/mesh-dns.md)
- [Configuration](docs/configuration.md)
- [Configuration reference](docs/configuration-reference.md)
- [Advanced peerings](docs/advanced-peerings.md)
- [FAQ](docs/faq.md)
- [Private network: two people only](docs/private-two-nodes.md)
- [Uninstall completely](docs/uninstall.md)
- [Key rotation after a leak](docs/key-rotation.md)
- [Testing and release checks](docs/TESTING.md)

### Installation guides

- [Ubiquiti EdgeOS / vyatta-uqda](docs/install-edgeos.md)
- [Windows](docs/install-windows.md)
- [macOS (manual build)](docs/install-macos-manual.md)
- [macOS (.pkg installer)](docs/install-macos-pkg.md)
- [Linux (manual build)](docs/install-linux-manual.md)
- [OpenWrt](docs/install-openwrt.md)
- [Gentoo](docs/install-gentoo.md)
- [Debian / Ubuntu / Mint](docs/install-debian.md)
- [RPM (Fedora / RHEL / CentOS)](docs/install-rpm.md)
- [FreeBSD](docs/install-freebsd.md)

## Building from source

```
./build
```

This produces `uqda` and `uqdactl` in the repository root.

## Configuration

Generate a new configuration:

```
./uqda -genconf > /path/to/uqda.conf
```

Run with a configuration file:

```
./uqda -useconffile /path/to/uqda.conf
```

Further options are documented in the repository and generated configuration comments.

## License

This repository’s `LICENSE` file is **GNU LGPLv3** (with the library linking exception in that file). The project source and layout follow that license; see `LICENSE` for the full text.
