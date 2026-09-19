# Uqda

[العربية](README.ar.md)

**Uqda Core is an independently maintained implementation compatible with the
existing Yggdrasil network, with security hardening and operational improvements.**
It provides encrypted IPv6 connectivity over IPv4 or IPv6 peer connections.
Uqda is derived from the open-source Yggdrasil implementation; it does not create
a separate network and does not imply upstream endorsement.

**Current development release: Uqda 26 Beta 1.** Beta 1 has not been published;
the intended release tag is `v26.0-beta.1`. See [CHANGELOG](CHANGELOG.md).

Uqda comes from **عُقدة**, meaning a knot, connection point or node. It is not
an acronym. A network node connects to other nodes and may forward traffic.

## Capabilities

- Public-key-derived IPv6 addresses and end-to-end encrypted sessions.
- TCP, TLS, QUIC, WebSocket, SOCKS and Unix-socket peer connectivity.
- Local multicast discovery and TUN integration.
- Local administration through `uqdactl` and the admin socket.
- Configuration validation with `-checkconf`, Unix key-permission warnings,
  non-local admin exposure warnings and jittered peer reconnect delays.

Routing and encrypted sessions build on Ironwood. Wire compatibility is a
requirement; the included two-process tests use Uqda at both ends and do not
replace independent upstream interoperability testing. See [compatibility](docs/compatibility.md).

## Installation

Install Go 1.25 or newer, then build from source:

```sh
git clone https://github.com/Uqda/Core.git
cd Core
./build
./uqda --version
./uqdactl version
```

On Windows use `build.bat`. The binaries are `uqda` and `uqdactl`. Packaging
sources target Debian, Windows MSI, macOS and Docker; native package validation
is required before distribution. See [installation](docs/installation.md).

## Quick start

For a **new** node, create a protected configuration file once:

```sh
umask 077
./uqda -genconf > uqda.conf
./uqda -useconffile uqda.conf -checkconf
./uqda -useconffile uqda.conf
```

Do not overwrite an existing configuration: it contains your persistent identity.
Add trusted peer URIs to `Peers`, or use multicast discovery on a trusted local
network. Creating the TUN interface normally requires administrator privileges;
Linux can use `CAP_NET_ADMIN`. See [configuration](docs/configuration.md).

```sh
./uqdactl getSelf
./uqdactl getPeers
```

`uqda -autoconf` uses a new random identity on each startup. For migration,
retain your existing key and follow [compatibility](docs/compatibility.md).

## Security and documentation

Uqda does not provide anonymity. The admin API has no authentication: keep it
local and restrict access. Applications on the node still need their own security.
Report vulnerabilities privately using [SECURITY](SECURITY.md).

- [Architecture](docs/architecture.md)
- [Configuration](docs/configuration.md) and [installation](docs/installation.md)
- [Compatibility](docs/compatibility.md) and [mobile API](docs/mobile.md)
- [Security model](docs/security.md)
- [Testing](docs/testing.md) and [contributing](CONTRIBUTING.md)
- [Packaging](docs/packaging.md) and [release conventions](docs/releasing.md)

## License and attribution

Uqda Core is licensed under LGPLv3 with the linking exception in [LICENSE](LICENSE).
See [NOTICE](NOTICE.md) for Yggdrasil, Ironwood and other upstream attribution.
