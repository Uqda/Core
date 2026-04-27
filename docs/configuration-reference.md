# Configuration reference

For a **narrative explanation** of what the network provides and what **every** option does in prose (plus URI parameters and `uqda` / `uqdactl` surfaces), see **[Network capabilities and complete configuration guide](network-services-and-config.md)**.

All options go in your **uqda.conf** (HJSON or JSON format).

Generate a default config with:

```sh
uqda -genconf > /etc/uqda.conf
```

Defaults for **AdminListen**, **IfName**, **IfMTU**, and **MulticastInterfaces** depend on the build target; see `src/config/defaults_*.go` in this repository.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| PrivateKey | hex string | auto-generated | Your node's ed25519 private key (hex-encoded in JSON). **Do not share.** |
| PrivateKeyPath | string | *(empty)* | Path to a PEM private key file. When set, **PrivateKey** in the file is ignored and the key is loaded from disk. |
| Certificate | *(not in file)* | runtime-generated | Self-signed TLS certificate derived from **PrivateKey**. Populated after load; omitted from JSON (`json:"-"`). |
| Peers | []string | `[]` | Outbound peer URIs (e.g. `tls://1.2.3.4:12345`). |
| InterfacePeers | map[string][]string | `{}` | Outbound peers keyed by source interface name (multi-homed setups). |
| Listen | []string | `[]` | Inbound listener URIs (e.g. `tls://[::]:12345`). Empty means no extra listeners beyond multicast-assisted discovery. |
| AdminListen | string | platform default | Admin socket for **uqdactl**. Examples: `unix:///var/run/uqda.sock` (Linux and others), `tcp://localhost:9001` (Windows / generic). Use `"none"` to disable. |
| MulticastInterfaces | []object | platform list | Multicast peer discovery entries. See [MulticastInterfaces options](#multicastinterfaces-options). |
| AllowedPublicKeys | []string | `[]` | Hex public keys allowed for **incoming** peerings. Empty = no restriction on incoming keys (still subject to **Listen** / multicast). **Not a firewall** for IPv6 traffic to your host. |
| IfName | string | platform default | TUN device name: `"auto"`, a fixed name (e.g. `uqda0`), `"/dev/tun0"` (FreeBSD), `"Uqda"` (Windows), or `"none"` to run without TUN. |
| IfMTU | uint64 | platform max | TUN MTU. Minimum **1280**. Upper bound is platform-specific (e.g. up to **65535** on Linux; lower on some BSDs). |
| LogLookups | bool | `false` | When `true`, registers an admin **lookups** handler and logs ironwood lookup debug data (development/diagnostics). |
| NodeInfoPrivacy | bool | `false` | When `true`, omits default nodeinfo fields (platform, arch, build version); only **NodeInfo** keys you set are exposed. |
| NodeInfo | map | `null` | Optional arbitrary key/value map visible to the network on request. |

## MulticastInterfaces options

Each element of **MulticastInterfaces** is an object with these fields (see `MulticastInterfaceConfig` in `src/config/config.go`):

| Option | Type | Description |
|--------|------|-------------|
| Regex | string | Regular expression matched against OS interface names; first matching block wins. |
| Beacon | bool | When `true`, advertise presence so other nodes can discover you. |
| Listen | bool | When `true`, listen for other beacons and attempt peerings. |
| Port | uint16 | TCP port for the link-local listener; **0** means choose automatically (omitted in JSON when zero). |
| Priority | uint64 | Relative priority for peerings on this interface vs others to the **same** node (lower is preferred; effectively **0–254**). |
| Password | string | Optional shared secret for authenticated multicast discovery (**max 64** characters). |

## Peer URI formats

Examples supported in **Peers**, **InterfacePeers**, and **Listen** (see **`uqda -genconf`** comments and `src/config` for full query parameters):

```
tcp://host:port
tls://host:port
tls://host:port?sni=domain.com
tls://host:port?password=secret
tls://host:port?priority=1
tls://host:port?maxbackoff=1h
quic://host:port
ws://host:port
wss://host:port
socks://proxy:port/host:port
sockstls://proxy:port/host:port
unix:///path/to/socket.sock
```

## See also

- [Network capabilities & full config guide](network-services-and-config.md) — narrative for every option and URI parameter.
- [Configuration](configuration.md) — modes, `genconf`, normalisation, keys.
- [Advanced peerings](advanced-peerings.md) — multi-homed, priorities, Tor, firewalls.
