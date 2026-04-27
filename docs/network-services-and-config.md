# Network capabilities and complete configuration guide

This document answers two questions together:

1. **What does the Uqda overlay “provide”** — what you can do on the network once nodes peer.  
2. **What every configuration option controls** in **`uqda.conf`** (and related CLI flags on **`uqda`**).

For a compact **table** of the same options, keep **[Configuration reference](configuration-reference.md)** open alongside this page.

---

## Part A — Capabilities the mesh offers (conceptual “services”)

Uqda is **not** a SaaS product with separate named cloud services. It is **one daemon** that, when configured and peered, gives you:

| Capability | What it means in practice |
|--------------|---------------------------|
| **Overlay IPv6 (`200::/7`)** | Every node gets a **stable address** derived from its key. Applications speak normal IPv6 over **TUN** as if it were another NIC. |
| **End-to-end encryption** | Payload between overlay addresses is protected inside the **Ironwood** encrypted packet layer; underlay links use **TLS** (or other schemes) for peering transports. |
| **Multi-hop routing** | Packets can traverse **intermediate Uqda nodes** toward a destination when direct peerings do not exist — behaviour is **shortest-path oriented** (see [About Uqda](about.md)). |
| **Decentralised operation** | No central allocator assigns your address; identity is **local key material**. Routing state spreads among peers. |
| **Outbound peerings (`Peers` / `InterfacePeers`)** | Your node **actively dials** other nodes over the Internet or LAN using the URI you configure. |
| **Inbound peerings (`Listen`)** | Your node **accepts** dial-ins on a fixed TCP (etc.) port — needed when the other side cannot dial you (e.g. you have a public IP / port forward). |
| **LAN discovery (`MulticastInterfaces`)** | On the same broadcast domain, compatible nodes can **find each other** via IPv6 multicast beacons and open **TLS** peerings automatically (optional password). |
| **Inbound key filter (`AllowedPublicKeys`)** | Limits **which remote public keys** may complete an **inbound** peering handshake to you. **Does not** replace a host firewall; **does not** block multicast-discovered peerings the same way — see [Private two-node](private-two-nodes.md). |
| **Routed `/64` per node** | Besides the host address, you get a **`300:…::/64`**-style prefix for downstream subnets (see [Configuration — advertising a prefix](configuration.md#advertising-a-prefix)). |
| **Optional metadata (`NodeInfo` / `NodeInfoPrivacy`)** | A small **queryable map** other nodes can read — useful for labels or diagnostics, **not** a general database. |
| **Operations API (`AdminListen` + `uqdactl`)** | Local (or configured TCP) **control plane**: inspect peers, tree, paths, sessions, TUN, multicast; add/remove peers at runtime. **Not authenticated** by default — treat like root access. |
| **Diagnostics (`LogLookups`)** | When enabled, extra **lookup logging** and admin handler for development/troubleshooting of routing lookups. |

**What Uqda does *not* include by itself:** DNS for the mesh (see **[Unified DNS for your mesh](mesh-dns.md)**), DHCP, NAT “exit to clearnet”, certificate authority PKI for users, or anonymity — combine with other tools if you need those ([FAQ](faq.md)).

---

## Part B — Every `uqda.conf` field explained

### Identity and cryptography

| Field | Role |
|--------|------|
| **`PrivateKey`** | Ed25519 private key (hex in file). **Secret.** Determines your overlay IPv6 and TLS identity. If it leaks, rotate — see [Key rotation](key-rotation.md). |
| **`PrivateKeyPath`** | If set, the key is loaded from a **PEM file** on disk and the inline **`PrivateKey`** field in the same file is **ignored**. Useful to keep keys out of the main config file. |
| **`Certificate`** | Not stored in the file: built at runtime from your key and used for **TLS** between peers. |

### Who connects to whom (underlay)

| Field | Role |
|--------|------|
| **`Peers`** | List of **outbound** URIs. At startup (and after admin adds), Uqda **dials** these addresses continuously (with backoff on failure). |
| **`InterfacePeers`** | Same as **`Peers`**, but each list is bound to a **source interface name** (e.g. `"eth0"`) for **multi-homed** routing policy. |
| **`Listen`** | List of **inbound** listener URIs. Without at least one listener (or multicast on LAN), others **cannot** dial you unless you dial them first. |

### Local integration (TUN)

| Field | Role |
|--------|------|
| **`IfName`** | TUN adapter name: fixed string, `"auto"`, OS-specific (`"Uqda"` on Windows, `"/dev/tun0"` on some BSDs), or **`"none"`** to run **without** a TUN interface (peering/router-only style use cases). |
| **`IfMTU`** | Maximum packet size on TUN. **Minimum 1280** (IPv6 requirement). Upper limit depends on OS. |

### Discovery and trust

| Field | Role |
|--------|------|
| **`MulticastInterfaces`** | Array of rules. Each rule matches OS interfaces by **`Regex`**, and sets **`Beacon`** (advertise), **`Listen`** (react to others), optional fixed **`Port`**, **`Priority`** between interfaces to the **same** remote node, and optional **`Password`** for beacon authentication. **Empty array `[]`** disables multicast entirely. |
| **`AllowedPublicKeys`** | Hex-encoded Ed25519 **public keys** allowed for **incoming** peering connections. Empty = no key-based restriction on inbound handshakes. |

### Control plane and metadata

| Field | Role |
|--------|------|
| **`AdminListen`** | Where **`uqdactl`** connects: e.g. **`unix:///var/run/uqda.sock`** or **`tcp://127.0.0.1:9001`**. Use **`"none"`** to disable admin (stronger lock-down; you lose `uqdactl` unless re-enabled). |
| **`NodeInfo`** | Arbitrary JSON-like map exposed to the network on request (strings, numbers, nested objects as supported). |
| **`NodeInfoPrivacy`** | If **`true`**, default built-in metadata (platform, arch, version) is **hidden**; only keys you set under **`NodeInfo`** appear. |
| **`LogLookups`** | If **`true`**, enables extra **ironwood lookup** logging and an admin **lookups** path for deep routing diagnosis (development / heavy debugging). |

---

## Part C — Peer / listener URI schemes and query parameters

### Schemes (underlay)

Supported in **`Peers`**, **`InterfacePeers`**, and **`Listen`** (see `src/core/link.go`):

| Scheme | Typical use |
|--------|-------------|
| **`tls://`** | Encrypted TCP; **default** recommendation for Internet peerings. |
| **`tcp://`** | Plain TCP (no extra TLS wrapper from Uqda’s link layer beyond protocol rules). |
| **`quic://`** | QUIC transport. |
| **`ws://` / `wss://`** | WebSocket / secure WebSocket (useful behind some HTTP proxies or CDNs). |
| **`unix://`** | Unix domain socket (local IPC style peerings). |
| **`socks://` / `sockstls://`** | Outbound peerings via SOCKS5 proxy (Tor-friendly patterns in [Advanced peerings](advanced-peerings.md)). |

### Query parameters (same grammar on dial and listen where applicable)

| Parameter | Meaning |
|-----------|---------|
| **`password=`** | Shared secret for the **link** (length capped; see code). Must **match** on both sides of that link if used. |
| **`priority=`** | Unsigned byte: when multiple peerings exist to the **same** remote node, **lower** value is preferred for traffic steering between those links. |
| **`maxbackoff=`** | Go **duration** string (e.g. `30m`, `1h`). Caps reconnect backoff; must be **≥ 5s**. |
| **`sni=`** | TLS **Server Name Indication** hostname when the host part of the URI is an IP literal but the certificate expects a name. |
| **`key=`** (repeatable) | Hex **Ed25519 public key** pins for the remote identity on this link. Mismatch fails the peering. |

Examples appear in **[Configuration reference](configuration-reference.md#peer-uri-formats)** and in **`uqda -genconf`** comments.

---

## Part D — `uqda` command-line flags tied to configuration

These do **not** live inside **`uqda.conf`** but are how you **produce, inspect, or consume** config:

| Flag | Purpose |
|------|---------|
| **`-genconf`** | Print a new template (with comments in HJSON). |
| **`-genconf` `-json`** | Same as JSON. |
| **`-useconffile PATH`** | Run using that file. |
| **`-useconf`** | Read config from **stdin**. |
| **`-autoconf`** | Run without a static file: ephemeral keys and LAN-oriented behaviour. |
| **`-normaliseconf`** | With `-useconf` / `-useconffile`, print config merged with defaults (optional **`-json`**). |
| **`-exportkey`** | With `-useconf` / `-useconffile`, write **PEM** private key to stdout for **`PrivateKeyPath`** workflows. |
| **`-address`**, **`-subnet`**, **`-publickey`** | With `-useconf` / `-useconffile`, print derived overlay address, routed subnet, or public key **without** starting full routing (useful for scripts). |
| **`-logto`**, **`-loglevel`** | Logging destination and verbosity. |
| **`-user`** | On Unix, drop privileges after opening listeners (see `cmd/uqda`). |
| **`-notifyfd`** | Signal readiness to **systemd**-style supervisors. |

---

## Part E — `uqdactl` operations (control “services”)

All require a running **`uqda`** and a reachable **`AdminListen`**:

| Command | Purpose |
|---------|---------|
| **`list`** | Lists supported admin commands. |
| **`getSelf`** | Your keys, addresses, build info. |
| **`getPeers`** | Direct peerings and stats. |
| **`getTree`** | Spanning tree / routing view. |
| **`getPaths`** | Path table snapshot. |
| **`getSessions`** | Session-oriented view. |
| **`getNodeInfo`** | Query **NodeInfo** for a key. |
| **`getMulticastInterfaces`** | Multicast state. |
| **`getTun`** | TUN configuration snapshot. |
| **`addPeer` / `removePeer`** | Runtime change to peer list (URI syntax same as config). |

Use **`-endpoint=`** if admin is not at the default.

---

## See also

- [Hosting a website on Uqda](hosting-on-uqda.md) — bind a normal web server to your overlay IPv6; firewall; DNS; TLS.  
- [Configuration reference](configuration-reference.md) — quick lookup table.  
- [Configuration](configuration.md) — workflows (`genconf`, normalise, export).  
- [Advanced peerings](advanced-peerings.md) — multi-homed, Tor, multicast firewall.  
- [Private two-node](private-two-nodes.md) — closed pair topology.  
- [FAQ](faq.md) — threat model and limits.  
- [Documentation hub](README.md) — full project map.
