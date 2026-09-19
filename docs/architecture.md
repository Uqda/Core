# Uqda Core architecture

Uqda Core provides encrypted IPv6 connectivity on the Yggdrasil network.
Its Go module is `github.com/Uqda/Core`.

## Components

| Component | Source | Responsibility |
|---|---|---|
| Daemon | `cmd/uqda` | Configuration, startup, privileges, signals and shutdown |
| Control CLI | `cmd/uqdactl` | Local admin API client |
| Configuration | `src/config` | HJSON/JSON, platform defaults, keys and certificates |
| Identity | `src/identity`, `src/core` | Ed25519 identity and TLS configuration |
| Addresses | `src/address` | Public-key-derived IPv6 addresses and subnets |
| Peers and transports | `src/core/link*.go` | Peer registration, dialing, listeners and reconnects |
| Routing and sessions | `github.com/Arceliar/ironwood` | Routing and end-to-end encrypted sessions |
| Discovery | `src/multicast` | Local multicast advertisements and peer discovery |
| Host networking | `src/tun`, `src/ipv6rwc` | TUN interface and IPv6 packet adaptation |
| Administration | `src/admin` | Unix/TCP socket and command handlers |
| Version | `src/version` | Product identity and machine version |
| Mobile | `contrib/mobile` | Exported `Uqda` binding type |

## Node and packet lifecycle

The daemon loads configuration and identity, creates the Core and its Ironwood
packet connection, configures peer listeners and outbound peers, and attaches
administration, multicast discovery and the TUN adapter. Signals initiate
shutdown of these components. The private key determines the node's address;
replacing it changes identity. Automatic configuration creates an ephemeral key.

Local IPv6 packets pass from TUN through `ipv6rwc` into Core and Ironwood's
encrypted sessions. The reverse path delivers decrypted packets to the host.
Transport links carry routed traffic; relay nodes forward encrypted session data.

## Peer lifecycle

The link manager owns peer registration, dialing, listeners and cancellation.
TCP, TLS, QUIC, WebSocket, secure WebSocket, SOCKS and Unix sockets have
transport-specific implementations. Handshake metadata is parsed before a
connection is handed to Ironwood. Reconnect delays include jitter. Removing a
peer synchronously removes its registration so an immediate re-add is possible.
Actor coordination uses `github.com/Arceliar/phony`.

## Security boundaries

Remote handshake data, discovery packets and decrypted IPv6 payloads are
untrusted. The admin API relies on socket placement and OS permissions; it
has no authentication layer. Configuration and PEM key files hold persistent
identity. See [security](security.md) for controls and limitations and
[compatibility](compatibility.md) for protocol boundaries.

## Build and platform integration

`build` and `build.bat` produce `uqda` and `uqdactl`. Packaging and service
integration live in `contrib/`; CI lives in `.github/workflows/`. `misc/`
contains manual network-namespace experiments. Dependency versions are pinned
in `go.mod` and `go.sum`. Go 1.25 or newer is required by this module.
There is no built-in automatic updater. Attribution is in [NOTICE](../NOTICE.md).
