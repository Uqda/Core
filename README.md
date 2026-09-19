# Uqda

[![Build status](https://github.com/Uqda/Core/actions/workflows/ci.yml/badge.svg)](https://github.com/Uqda/Core/actions/workflows/ci.yml)

اقرأ هذا الملف بالعربية: [README.ar.md](README.ar.md)

**Uqda Core is an independently maintained, hardened and modernized
implementation compatible with the existing Yggdrasil network.**

Uqda does not create a new network or a new protocol. It speaks the same
Yggdrasil wire protocol as the original implementation and participates on
the same, existing Yggdrasil network - a Uqda node and an upstream
Yggdrasil node can peer with each other directly, with no configuration
workaround required. What Uqda changes is the implementation itself:
engineering quality, security hardening, reliability, diagnostics,
packaging, and long-term maintainability. See [NAMING.md](NAMING.md) for
the full naming policy this project holds itself to, and
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for exactly
what has and hasn't been done yet.

This project originated as a fork of `yggdrasil-go`
(github.com/yggdrasil-network/yggdrasil-go) and remains derived from it -
see [NOTICE.md](NOTICE.md) for upstream attribution. Uqda is independently
maintained; nothing here implies endorsement by, or affiliation with, the
Yggdrasil project's original authors or maintainers.

## About the name

**Uqda** comes from the Arabic word **عُقدة**, meaning a knot, a
connection point, or a node. In networking, a node is a participating
point in a network - it connects to other nodes and may help forward
traffic for the network as a whole. Uqda is not an acronym.

## Introduction

Yggdrasil - the network Uqda participates in - is an early-stage
implementation of a fully end-to-end encrypted IPv6 network. It is
lightweight, self-arranging, supported on multiple platforms and allows
pretty much any IPv6-capable application to communicate securely with
other nodes on the network. It does not require you to have IPv6 Internet
connectivity - it also works over IPv4.

## Supported Platforms

Uqda Core targets Linux, macOS, Windows, FreeBSD and OpenBSD. See
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md) for exactly
which platforms have actually been build-tested versus just targeted, and
`contrib/` for platform-specific packaging and scripts.

## Building

If you want to build from source, as opposed to installing one of the
pre-built packages:

1. Install [Go](https://golang.org) (requires Go 1.22 or later)
2. Clone this repository
3. Run `./build`

This produces two binaries: `uqda` (the daemon) and `uqdactl` (the local
control CLI). Note that you can cross-compile for other platforms and
architectures by specifying the `GOOS` and `GOARCH` environment variables,
e.g. `GOOS=windows ./build` or `GOOS=linux GOARCH=mipsle ./build`.

## Running

### Generate configuration

To generate static configuration, either generate a HJSON file (human-friendly,
complete with comments):

```
./uqda -genconf > /path/to/uqda.conf
```

... or generate a plain JSON file (which is easy to manipulate
programmatically):

```
./uqda -genconf -json > /path/to/uqda.conf
```

You will need to edit the `uqda.conf` file to add or remove peers, modify
other configuration such as listen addresses or multicast addresses, etc.
You can validate a configuration file without starting the node using
`-checkconf`:

```
./uqda -useconffile /path/to/uqda.conf -checkconf
```

### Run Uqda

To run with the generated static configuration:

```
./uqda -useconffile /path/to/uqda.conf
```

To run in auto-configuration mode (which will use sane defaults and random keys
at each startup, instead of using a static configuration file):

```
./uqda -autoconf
```

You will likely need to run Uqda as a privileged user or under `sudo`,
unless you have permission to create TUN/TAP adapters. On Linux this can be done
by giving the `uqda` binary the `CAP_NET_ADMIN` capability.

Use `uqdactl` to inspect or control a running node locally, e.g.
`uqdactl getSelf` or `uqdactl getPeers`.

## Documentation

Uqda-specific documentation lives in this repository:
[ARCHITECTURE.md](ARCHITECTURE.md), [SECURITY.md](SECURITY.md),
[docs/THREAT_MODEL.md](docs/THREAT_MODEL.md), and
[docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md).

Because Uqda preserves the Yggdrasil wire protocol and configuration
format unchanged, the upstream Yggdrasil project's own reference
documentation for peer URIs, multicast configuration, and general network
concepts remains accurate and directly applicable:

- [Configuration reference](https://yggdrasil-network.github.io/configurationref.html)
- [Yggdrasil network FAQ](https://yggdrasil-network.github.io/faq.html)
- [Upstream version history](CHANGELOG.md) (inherited from `yggdrasil-go`; see that file's own entries for their original dates and authorship)

## Communities

Because Uqda nodes participate on the same Yggdrasil network, the
existing Yggdrasil community resources are also relevant to Uqda
operators: the `#yggdrasil` IRC channel on [libera.chat](https://libera.chat)
and various others on [Yggdrasil-internal IRC networks](https://yggdrasil-network.github.io/services.html#irc).
These are upstream community spaces, not run by this project.

## License

This code is released under the terms of the LGPLv3, but with an added exception
that was shamelessly taken from [godeb](https://github.com/niemeyer/godeb).
Under certain circumstances, this exception permits distribution of binaries
that are (statically or dynamically) linked with this code, without requiring
the distribution of Minimal Corresponding Source or Minimal Application Code.
For more details, see: [LICENSE](LICENSE). For upstream and third-party
attribution (Yggdrasil, Ironwood, and others this project depends on), see
[NOTICE.md](NOTICE.md).
