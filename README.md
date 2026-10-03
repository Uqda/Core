# Uqda

[العربية](README.ar.md)

**Uqda Core is an independently maintained implementation compatible with the
existing Yggdrasil network, with security hardening and operational improvements.**
It provides encrypted IPv6 connectivity over IPv4 or IPv6 peer connections.
Uqda is derived from the open-source Yggdrasil implementation; it does not create
a separate network and does not imply upstream endorsement.

**Current release: Uqda 26.0.2 (`v26.0.2`).** It is ready for public use and testing;
report any problems through the
[bug report form](https://github.com/Uqda/Core/issues/new?template=bug_report.yml).
See [CHANGELOG](CHANGELOG.md) for the tested scope and known limitations.

Uqda comes from **عُقدة**, meaning a knot, connection point or node. It is not
an acronym. A network node connects to other nodes and may forward traffic.

## Capabilities

- Public-key-derived IPv6 addresses and end-to-end encrypted sessions.
- TCP, TLS, QUIC, WebSocket, SOCKS and Unix-socket peer connectivity.
- Local multicast discovery and TUN integration.
- Local administration through `uqdactl` and the admin socket.
- One-command local health checks with `uqdactl doctor` (or `uqdactl status`).
- Active remote reachability and loss checks with `uqdactl test <Uqda IPv6>`.
- Configuration validation with `-checkconf`, Unix key-permission warnings,
  non-local admin exposure warnings and jittered peer reconnect delays.

Routing and encrypted sessions build on Ironwood. Wire compatibility is a
requirement; the independent pinned-upstream gate checks real daemon peering and IPv6 packet
delivery over TCP/TLS, including reconnect and multi-hop forwarding. See [compatibility](docs/compatibility.md).

## Installation

### Using your Umbrel services

Uqda can provide an encrypted IPv6 path between your Umbrel host and another
Uqda-connected device. The useful outcome is access to an existing service,
such as your files or SSH, after configuring that service's IPv6 listening,
authentication and firewall. A connected peer alone is not proof that an app
is reachable. Apps are not published automatically, and opening the dashboard
on a phone does not connect the phone to the overlay.
See the [Umbrel service-access guide](docs/umbrel.md#your-services-over-uqda)
for the deliberate setup flow and the status of the new address-book workflow.

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

Use the compact terminal dashboard: `uqda`, `uqda peers`, `uqda info` and
`uqda test <IPv6>`. The controller commands below remain supported for advanced
administration. See the [command guide](docs/commands.md).

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
./uqdactl doctor
./uqdactl test 200:1234::1
```

The doctor check explains daemon, identity, peer and TUN status without changing
configuration or showing peer passwords. See [doctor](docs/doctor.md).

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
- [Doctor health check](docs/doctor.md)
- [Testing](docs/testing.md) and [contributing](CONTRIBUTING.md)
- [Packaging](docs/packaging.md) and [release conventions](docs/releasing.md)

## License and attribution

Uqda Core is licensed under LGPLv3 with the linking exception in [LICENSE](LICENSE).
See [NOTICE](NOTICE.md) for Yggdrasil, Ironwood and other upstream attribution.
